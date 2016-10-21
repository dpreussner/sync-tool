# A filesync tool

A filesync tool which synchronises source directories with destination directories. Written in go.

## Misc

Working and quite reliable. Known limitation crashes when the source root folder is deleted.

## Example usage

    //silent
    ./synctool -c path/to/config.json

    //verbose
    ./synctool -v -c path/to/config.json



## Command line options

    -h         Show help/available options
    -c *path*  The path to a valid filesync JSON config file
    -v         Verbose, loggs every action to stdout
    -no-clean  Set if no clean on startup should be done
    -o         Set if the sync process should only be done once
    -tick      Set the duration (ms) between scans of the source folders

## Config attributes


    *mappings*              An array of mapping objects
        *srcRoot*           A path to source folder either relative to the sync.exe or absolute (Unix style paths)
        *dstRoot*           The target folder to which shall be synced. (Unix style paths)
        *files*             A pattern which has to be matched in order for a file to be synced.
        *ignored*           A pattern which has to be matched in order for a file to not be synced.
        *cleanupPatterns*   A pattern which is used for cleanup of folders or files to be deleted before the sync process starts.

## Example config file:

    {
        "mappings": [
            {
                "srcRoot": "src/main/resources",
                "dstRoot": "C:/myfolder/test",
                "files": "**/*.xml",                // copy all XML files from srcRoot to dstRoot
                "ignored": "*.xsd",                 // ignore XSDs
                "cleanupPatterns": [
                    "**/test",                      // delete all folders in dstRoot which are named test on startup                         
                    "**/*.xml"                      // delete all xml files on startup
                ]

            }
        ]
    }


