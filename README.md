# patch_installer

## NAME
`patch_installer` - A tool for downloading and installing MSU patches, with optional ZIP extraction.

## DESCRIPTION
The `patch_installer` program downloads and installs MSU patches from provided links. It supports downloading ZIP files, extracting their contents, and installing the MSU files within. Logging is provided for all actions taken during the process.

## OPTIONS
- `-silent`  
  Run the program in silent mode. Requires the `-links` parameter to be specified.
- `-links`  
  Specify the file containing the download links. Each link should be on a new line.


## INTERACTIVE MODE
When run without the `-silent` option, the program will prompt the user to enter 'zip' for downloading a ZIP file or 'msu' for downloading a direct MSU file. The user will then be prompted to enter the download link.

## FILES
The following files are created and used by the program:
- `patch_installation_log.txt`  
  Log file containing all actions taken during the execution of the program.
- `c:\temp\patchinstalls`  
  Directory where downloaded files and extracted contents are stored.

## USAGE EXAMPLES
Run the program in silent mode with a links file:
`patch_installer.exe -silent -links links.txt`


Run the program interactively:
`patch_installer.exe`


## COPYRIGHT
This is free software; see the source for copying conditions. There is NO warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

