package main

import (
	"dblist"
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

	delArchived := flag.Bool("delete_archived", false, "deletes files with attribute archived set")
	configfile := flag.String("config", "", "full config file name (lists databases files groups)")
	flag.Parse()
	if *configfile == "" {
		flag.Usage()
		os.Exit(1)
	}
	conf := dblist.ReadConfig(*configfile)

	unique := dblist.GetUniquePaths(conf)

	files := dblist.ReadFilesFromPaths(unique)

	lastfilesmap := make(map[string][]dblist.FileInfoWin)

	if *delArchived {
		for dir, slice := range files {
			lastfilesslice := dblist.GetLastFilesGroupedByFunc(slice, dblist.GroupFunc)
			// make lastfilesslice descending to use sort.Search
			sort.Slice(lastfilesslice, func(i, j int) bool {
				return lastfilesslice[i].Name() > lastfilesslice[j].Name() //DESC
			})
			lastfilesmap[dir] = lastfilesslice
		}
		deleteArchivedFiles(files, lastfilesmap)
	} else {
		flag.Usage()
	}

}

// deleteArchivedFiles needs exceptfiles be sorted descending
func deleteArchivedFiles(files, exceptfiles map[string][]dblist.FileInfoWin) {
	for dir, slice := range files { // for every dir
		for _, finf := range slice { // for every file in dir
			if !finf.IsDir() { // dont touch subdirs and not archived
				if (finf.WinAttr & syscall.FILE_ATTRIBUTE_ARCHIVE) == 0 { // no ARCHIVE means file has been archived

					// search if file in question is in the exception list
					curmap := exceptfiles[dir]                         // exceptfiles is descending
					pos := sort.Search(len(curmap), func(i int) bool { // bin search
						return curmap[i].Name() <= finf.Name() // <= for the descending slice

					})
					if pos < len(curmap) && curmap[pos].Name() == finf.Name() {
						// found
						continue // file should not be deleted. this is the last file in the group
					}
					fullFilename := filepath.Join(dir, finf.Name())
					log.Printf("deleting file %s\n", fullFilename)
					if true { //DEBUG
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
