package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"slices"
)

var imdbmap = make(map[string][]MovieData)
var tmdbmap = make(map[string][]MovieData)

var vidXtns = []string{".avi", ".mpg", ".mkv", ".mp4", ".m4v", ".wmv", ".ts", ".wmtv", ".ogv", ".m4v", ".wtv", ".flv", ".mov", ".dvr-ms", ".iso"}

type MovieData struct {
	Status        xml.Name `xml:"movie"`
	Title         string   `xml:"title"`
	Year          string   `xml:"year"`
	OriginalTitle string   `xml:"originaltitle"`
	DateAdded     string   `xml:"dateadded"`
	IMDBid        string   `xml:"imdbid"`
	TMDBid        string   `xml:"tmdbid"`
	Path          string
}

func rememberMovie(a *MovieData, key string, historyMap map[string][]MovieData) {
	if key != "" {
		value, known := historyMap[key]
		if !known {
			historyMap[key] = []MovieData{*a}
		} else {
			historyMap[key] = append(value, *a)
		}
	}
}

func walkNfos(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	//if this is a nfo.. read it to get the id.
	if strings.HasSuffix(path, ".nfo") {

		dirname := filepath.Dir(path)
		//skip nfo's in bdmv dirs
		if filepath.Base(path) == "index.nfo" && filepath.Base(dirname) == "BDMV" {
			return nil
		}

		// Open our xmlFile
		xmlFile, err := os.Open(path)
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println("Unable to open " + path)
			return err
		}

		// defer the closing of our xmlFile so that we can parse it later on
		defer xmlFile.Close()

		// read our opened xmlFile as a byte array.
		byteValue, err := io.ReadAll(xmlFile)
		if err != nil {
			fmt.Println("Couldn't read xml data into buffer " + path)
			return err
		}

		var data = &MovieData{}
		if err := xml.Unmarshal(byteValue, data); err != nil {
			fmt.Println("Unable to parse file content as xml " + path)
			return nil
		}

		data.Path = path

		if data.IMDBid != "" {
			rememberMovie(data, data.IMDBid, imdbmap)
		} else {
			if data.TMDBid != "" {
				rememberMovie(data, data.TMDBid, tmdbmap)
			}
		}

	}
	return nil
}

func walkMovies(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	//if this is a nfo.. read it to get the id.
	if strings.HasSuffix(path, ".nfo") {

		dirname := filepath.Dir(path)

		//skip nfo's in bdmv dirs
		if filepath.Base(path) == "index.nfo" && filepath.Base(dirname) == "BDMV" {
			return nil
		}

		entries, err := os.ReadDir(dirname)
		if err != nil {
			log.Fatal(err)
		}

		found := false
		for _, e := range entries {

			xtn := filepath.Ext(e.Name())

			if slices.Contains(vidXtns, strings.ToLower(xtn)) {

				//it's a movie file.. is it the right one tho?
				filename := filepath.Base(path)
				filename = strings.TrimSuffix(filename, ".nfo")
				moviename := strings.TrimSuffix(e.Name(), xtn)

				found = filename == moviename
				if found {
					break
				}
			}
		}

		if !found {
			//last chance.. is there a subdir named BDMV ? .. if so, it's fine.
			if _, err := os.Stat(filepath.Join(dirname, "BDMV")); os.IsNotExist(err) {
				fmt.Println(dirname + "  " + filepath.Base(path))
			}
		}

	}
	return nil
}

func testPath(path string) {
	fmt.Println("Loading data from " + path)
	err := filepath.Walk(path, walkNfos)
	if err != nil {
		log.Println(err)
	}
}

func listDirsWithNoMovie(path string) {
	fmt.Println("Checking for dirs with no video at " + path)
	err := filepath.Walk(path, walkMovies)
	if err != nil {
		log.Println(err)
	}
}

func guessName(filename string, folderName string) string {

	//origfilename := filename

	//remove nfo extension..
	filename = strings.TrimSuffix(filename, ".nfo")

	//trim some popular prefixes..
	filename = strings.TrimPrefix(filename, "TwoDDL_")
	filename = strings.TrimPrefix(filename, "HDPOPCORNS")
	filename = strings.TrimPrefix(filename, "[snahp.it]]")
	filename = strings.TrimPrefix(filename, "jauto_")

	fileRunes := []rune(filename)
	//try to figure out the part to trim from the start of the filename..
	//this is fuzzy, as there are too many different approaches..
	fileIdx := 0
	removePrefix := ""
	for _, r := range folderName {
		if !isLetterOrNumber(r) {
			continue
		}
		if isLetterOrNumber(r) {
			//skip chars in filename that are nonalphanum, and that don't match this char in foldername
			for !isLetterOrNumber(fileRunes[fileIdx]) || !caseInsensitiveAlphaNumMatch(r, fileRunes[fileIdx]) {
				removePrefix = removePrefix + string(fileRunes[fileIdx])
				//advance, if we can, exit if we cant
				fileIdx++
				if fileIdx == len(fileRunes) {
					break
				}
			}

			if fileIdx == len(fileRunes) {
				break
			}

			//matching char.. add to prefix
			removePrefix = removePrefix + string(fileRunes[fileIdx])
			fileIdx++
			if fileIdx == len(fileRunes) {
				break
			}
		}
	}

	//if we added the entire filename to remove, then use the entire filename as the tag
	altVersionTag := ""
	if len(removePrefix) == len(fileRunes) {
		altVersionTag = filename
	} else {
		//after removing the prefix, we are left with our altVersionTag that needs further cleanup.
		altVersionTag = strings.TrimPrefix(filename, removePrefix)
		removePrefix = ""
		for _, r := range altVersionTag {
			if !isLetterOrNumber(r) && !(r == '{' || r == '(') {
				removePrefix = removePrefix + string(r)
			} else {
				break
			}
		}
		altVersionTag = strings.TrimPrefix(altVersionTag, removePrefix)
	}

	//remove some popular suffixes in the altVersionTag to clean things up a little..
	altVersionTag = strings.TrimSuffix(altVersionTag, "www.tuserie.com")
	altVersionTag = strings.TrimSuffix(altVersionTag, "_snahp.it")
	altVersionTag = strings.TrimSuffix(altVersionTag, ".")

	fixed := folderName + " - " + altVersionTag + ".nfo"

	return fixed
}

func isLetterOrNumber(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

func caseInsensitiveAlphaNumMatch(a rune, b rune) bool {
	if unicode.IsNumber(a) && unicode.IsNumber(b) {
		return a == b
	}
	if unicode.IsLetter(a) && unicode.IsLetter(b) {
		return unicode.ToLower(a) == unicode.ToLower(b)
	}
	return false
}

func suggestRenames(current []MovieData, folderName string) {
	fmt.Println("Suggested renames:")
	for _, d := range current {
		filename := filepath.Base(d.Path)
		if !strings.HasPrefix(filename, folderName) {
			newName := guessName(filename, folderName)
			newName = strings.TrimSuffix(newName, ".nfo")
			filename = strings.TrimSuffix(filename, ".nfo")

			//minimum escapes for powershell, as we are using ' quoted strings, literal ' becomes ''
			newName = strings.ReplaceAll(newName, "'", "''")
			filename = strings.ReplaceAll(filename, "'", "''")
			dirname := strings.ReplaceAll(filepath.Dir(d.Path), "'", "''")

			//escape the dir name for cd using a suggested powershell built-in
			//escape the source name for replace using regex escape
			fmt.Println("$loc = [Management.Automation.WildcardPattern]::Escape('" + dirname + "') ; cd $loc ; Get-ChildItem -Filter '" + filename + "*.*' | Rename-Item -Newname { $_.Name -replace [regex]::escape('" + filename + "'),'" + newName + "'}")
		}
	}
}

func dumpDupes(historyMap map[string][]MovieData) {
	for _, val := range historyMap {
		//movie has multiple versions.. do they share a common folder?
		if len(val) > 1 {
			var firstdir = ""
			var dirsMismatch = false
			for idx, ver := range val {
				if idx == 0 {
					firstdir = filepath.Dir(ver.Path)
				} else {
					if firstdir != filepath.Dir(ver.Path) {
						dirsMismatch = true
						break
					}
				}
			}
			if dirsMismatch {
				fmt.Println("Multi folder Dupe:")
				for _, d := range val {
					fmt.Println(" " + d.Path)
				}
			} else {
				//dirs matched.. do all the nfo's share the dirname as a prefix?
				if true { //set to false disable multiver until dupes are sorted =)

					prefix := filepath.Base(firstdir) + " - "
					var prefixBad = false
					for _, d := range val {
						if !strings.HasPrefix(filepath.Base(d.Path), prefix) {
							prefixBad = true
							break
						}
					}
					if prefixBad {
						//dump the bad nfo paths
						fmt.Println("Bad Multi-Version content:")
						for _, d := range val {
							if !strings.HasPrefix(filepath.Base(d.Path), prefix) {
								fmt.Println(" " + d.Path)
							}
						}
						//dump suggested renames for this movie
						suggestRenames(val, filepath.Base(firstdir))
					}
				}
			}
		}
	}
}

func main() {

	//add additional lines here to check multiple dirs for .nfo files declaring a movie, but lacking a corresponding movie file.
	listDirsWithNoMovie(`\\server\\videos\\movies`)

	//add additional lines here to check multiple dirs and build an in memory table of movie file details by imdb/tmdb id.
	testPath(`\\server\\videos\\movies`)
	testPath(`\\server\videos\mpgs`)

	//scan the in memory table of movie file details, looking for movies that exist in multiple dirs, or movies within a single dir
	//with an incorrect prefix to be recognized as multi-version by emby.
	//where movies are in a single dir with bad names, powershell commands are generated to perform a suggested rename.
	//carefully review the suggested target filename (picking it is more an art than a science) and if acceptable, paste to powershell.
	fmt.Println("\nDupes by IMDB id:")
	dumpDupes(imdbmap)
	fmt.Println("\nDupes by TMDB id:")
	dumpDupes(tmdbmap)
}
