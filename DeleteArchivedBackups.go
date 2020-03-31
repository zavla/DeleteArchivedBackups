// package main leaves recent backups only. It doesn't delete last backup files.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/zavla/dblist/v3"
)

var (
	printExample    bool
	delArchived     bool
	dryRun          bool
	configfile      string
	keepLastNcopies uint
)

func init() {
	// register flags in init() allows me to go test the main package
	flag.BoolVar(&printExample, "example", false,
		"print example of config file")
	flag.BoolVar(&delArchived, "withArchiveAttr", false,
		"deletes files with attribute archived set")
	flag.StringVar(&configfile, "config", "",
		"config `file` name")
	flag.BoolVar(&dryRun, "dryrun", false,
		"print commands (doesn't actually delete files)")
	flag.UintVar(&keepLastNcopies, "keeplastN", 2,
		"keep recent N copies")

}

const exampleconf = `
[
	{"path":"./testdata/files", "Filename":"E08",    "suffix":"-FULL.bak",  "Days":2},
	{"path":"./testdata/files", "Filename":"A2",     "suffix":"-FULL.bak",  "Days":10}, 
	{"path":"./testdata/files", "Filename":"A2",     "suffix":"-differ.dif", "Days":1}, 
	{"path":"./testdata/files/bases116", "Filename":"dbase1", "suffix":"-FULL.bak",  "Days":5},
	{"path":"./testdata/files/bases116", "Filename":"dbase1", "suffix":"-differ.dif", "Days":1}
]
`

func main() {
	flag.Parse()

	//println(exampleconf)

	if printExample {
		fmt.Println(exampleconf)
		return
	}
	if configfile == "" {
		flag.Usage()
		os.Exit(1)
	}
	conf, err := dblist.ReadConfig(configfile)
	if err != nil {
		log.Printf("Config file read error: %s", err)
		return
	}

	// sort conf by databases names to allow sort.Search by database name
	sort.Slice(conf, func(i, j int) bool {
		return conf[i].Filename < conf[j].Filename
	})

	// get map with filename to suffixes
	uniquesuffixes := dblist.GetMapFilenameToSuffixes(conf)

	// gets unique folders names
	uniquedirs := dblist.GetUniquePaths(conf)

	// get existing files in each unique folder
	files := dblist.ReadFilesFromPaths(uniquedirs)

	lastfilesmap := make(map[string][]dblist.FileInfoWin)

	// Append to dontDeleteLastbackupfiles all files that should NOT be deleted:
	// the most recent files and files that have no config lines for them.
	for dir, filesindir := range files {
		// We get most recent files. The will not be deleted.
		dontDeleteLastbackupfiles := dblist.GetLastFilesGroupedByFunc(filesindir, dblist.GroupFunc, uniquesuffixes, keepLastNcopies)

		// A directory may contain some extra files not covered by config file - don't delete them.
		notinConfigFile := dblist.GetFilesNotCoveredByConfigFile(filesindir, conf, dblist.GroupFunc, uniquesuffixes)
		dontDeleteLastbackupfiles = append(dontDeleteLastbackupfiles, notinConfigFile...)

		// We exploit descending order while searching on the list of last files.
		sort.Slice(dontDeleteLastbackupfiles, func(i, j int) bool {
			return dontDeleteLastbackupfiles[i].Name() > dontDeleteLastbackupfiles[j].Name() //DESC
		})
		lastfilesmap[dir] = dontDeleteLastbackupfiles
	}
	deleteArchivedFiles(files, lastfilesmap, delArchived, dryRun)

}

// deleteArchivedFiles needs exceptfiles to be sorted descending.
// Exceptfiles sorted descending because it was used to get most recent backup file name earlier in the code.
func deleteArchivedFiles(files, exceptfiles map[string][]dblist.FileInfoWin, delArchived, dryrun bool) {
	for dir, slice := range files { // for every dir
		for _, finf := range slice { // for every file in dir
			if !finf.IsDir() { // dont touch subdirs and not archived
				// No ARCHIVE atrribute means that file has been archived or copied somewhere.
				// If file has Archive attribute you need a flag delArchived == true
				if (finf.WinAttr&syscall.FILE_ATTRIBUTE_ARCHIVE) == 0 ||
					delArchived { // or you insist on deleting files with Archive attr set

					// search if current file is in the exception list
					exeptionsForDir := exceptfiles[dir]                         // exceptfiles is already sorted descending
					pos := sort.Search(len(exeptionsForDir), func(i int) bool { // we use bin search
						return exeptionsForDir[i].Name() <= finf.Name() // <= for the descending slice

					})
					if pos < len(exeptionsForDir) && exeptionsForDir[pos].Name() == finf.Name() {
						// found
						continue // file should not be deleted. this is the last file in the group of backup files
					}
					fullFilename := filepath.Join(dir, finf.Name())
					log.Printf("rm %s\n", fullFilename)
					if !dryrun {
						err := os.Remove(fullFilename)
						if err != nil {
							log.Printf("%s\n", err)
						}
					}
				}
			}
		}
	}
}

//
