# dupefinder
A quick tool to scan directories to ensure filenames are appropriate for emby

Tool will look for 
 - Dirs containing .nfo files containing xml identifying a movie, but without a corresponding movie in the same dir.
   - eg, finds filename.nfo with valid xml, but no corresponding filename.mkv (or other extensions in code)
 - Multiple Dirs containing the same movie (as declared by xml in the .nfo)
 - Single Dirs containing multiple of the same movie, where the filename is not 'Directory Name - VersionIdentifier.mkv' 
   - eg. 'Top Gun (1986) - DVD.mkv'   (extension unimportant, must match the 'Directory Name - ' prefix)
   - Where multiple of the same movie are found in the same dir, where possible suggested renames are offered as powershell commands.
