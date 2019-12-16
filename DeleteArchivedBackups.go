package main

import (
	"dblist/v2"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"syscall"
)

func main() {

	// Example of config file:
	// [{"path":"g:/ShebB", "Filename":"buh_log8", "Days":1},
	// {"path":"g:/ShebB", "Filename":"buh_log3", "Days":1},
	// {"path":"g:/ShebB", "Filename":"buh_prom8", "Days":1},
	// ]

	delArchived := flag.Bool("withArchiveAttr", false, "deletes files with attribute archived set")
	configfile := flag.String("config", "", "full config file name (lists databases files groups)")
	dryRun := flag.Bool("dryrun", false, "dry run (doesn't actually delete files)")
	keepLastNcopies := flag.Uint("keeplastN", 2, "keep recent N copies")

	flag.Parse()
	if *configfile == "" {
		flag.Usage()
		os.Exit(1)
	}
	conf, err := dblist.ReadConfig(*configfile)
	if err != nil {
		log.Printf("Config file read error: %s", err)
		return
	}
	// sort conf by databases names to allow sort.Search by database name
	sort.Slice(conf, func(i, j int) bool {
		if conf[i].Filename < conf[j].Filename {
			return true
		}
		return false
	})

	// collect unique suffixes
	suffixes := make(map[string]int)
	for _, val := range conf {
		if _, has := suffixes[val.Suffix]; !has {
			suffixes[val.Suffix] = 0
		}

	}
	// transform a map into slice
	uniquesuffixes := make([]string, 0, len(suffixes))
	for suf := range suffixes {
		uniquesuffixes = append(uniquesuffixes, suf)
	}
	// gets unique folders names
	unique := dblist.GetUniquePaths(conf)

	// get files of each folder
	files := dblist.ReadFilesFromPaths(unique)

	lastfilesmap := make(map[string][]dblist.FileInfoWin)

	// We append to dontDeleteLastbackupfiles files that should not be deleted: last files and file that have no config line for them.
	for dir, filesindir := range files {
		// We get most recent files. The will not be deleted.
		dontDeleteLastbackupfiles := dblist.GetLastFilesGroupedByFunc(filesindir, dblist.GroupFunc, uniquesuffixes, *keepLastNcopies)

		notinConfigFile := dblist.GetFilesNotCoveredByConfigFile(filesindir, conf, dblist.GroupFunc, uniquesuffixes)

		// some actual files may not be in config file
		dontDeleteLastbackupfiles = append(dontDeleteLastbackupfiles, notinConfigFile...)

		// sort dontDeleteLastbackupfiles descending to use sort.Search later.
		// We exploite descending order to find most recent backup files.
		sort.Slice(dontDeleteLastbackupfiles, func(i, j int) bool {
			return dontDeleteLastbackupfiles[i].Name() > dontDeleteLastbackupfiles[j].Name() //DESC
		})
		lastfilesmap[dir] = dontDeleteLastbackupfiles
	}
	deleteArchivedFiles(files, lastfilesmap, *delArchived, *dryRun)

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
					curmap := exceptfiles[dir]                         // exceptfiles is already sorted descending
					pos := sort.Search(len(curmap), func(i int) bool { // we use bin search
						return curmap[i].Name() <= finf.Name() // <= for the descending slice

					})
					if pos < len(curmap) && curmap[pos].Name() == finf.Name() {
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
