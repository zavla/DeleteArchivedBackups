// package main rotates backup files - leaves recent backups only. It doesn't delete last backup files.
// Every db backup file may be a FULL db backup or a differential db backup.
// Never leave a differential backup without a correcponding FULL backup.
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
	logfile         string
)

var gitCommit string = "no version"

func init() {
	// registering program parameters-flags in init() allows to go test the 'main' package
	flag.BoolVar(&printExample, "example", false,
		"print example of config file config.json.")
	flag.BoolVar(&delArchived, "A", false,
		"also delete files that have archive attribute 'A'.")
	flag.StringVar(&configfile, "config", "",
		"config JSON `file` name, required.")
	flag.BoolVar(&dryRun, "dryrun", false,
		"print shell commands for deletion (doesn't actually delete files).")
	flag.UintVar(&keepLastNcopies, "keeplastN", 2,
		"keep recent N full copies for each database.")
	flag.StringVar(&logfile, "log", "stdout",
		"`log file` name. If you want to keep track of files that were deleted.")
}

//const exampleconf = "test string\r\n"

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
		fmt.Print(exampleconf)
		return
	}

	if len(os.Args) == 0 || configfile == "" {
		fmt.Printf(
			`
DeleteArchivedBackups.exe, ver. %v
			This is a tool for deletion of backup files of databases.
			Backup file names consist of database name and the datetime. 
			It does not delete N last copies of each database backups.
			It uses config JSON file to know database names.
Example: DeleteArchivedBackups.exe -A -config ./config.json
`, gitCommit)
		flag.Usage()
		os.Exit(1)
	}
	var logwriter *os.File
	if logfile != "std out" {
		logfile, err := filepath.Abs(logfile)
		if err != nil {
			fmt.Printf("Log file name incorrect: %v\n", err)
			os.Exit(1)
		}
		logwriter, err = os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.SetOutput(logwriter)
		defer func() {
			log.SetOutput(os.Stdout)
			logwriter.Close()
		}()

	}
	conf, err := dblist.ReadConfig(configfile)
	if err != nil {
		log.Printf("Config file read error: %s\n", err)
		return
	}
	if len(conf) == 0 {
		log.Printf("Config file %v is empty. Example:\n", configfile)
		log.Print(exampleconf)
		return
	}
	// sort conf slice by databases names to allow sort.Search by database name
	sort.Slice(conf, func(i, j int) bool {
		return conf[i].Filename < conf[j].Filename
	})

	// get map with filenames to files suffixes
	uniquesuffixes := dblist.GetMapFilenameToSuffixes(conf)

	// gets unique folders names with backups from config file
	uniquedirs := dblist.GetUniquePaths(conf)

	// get all existing files in each unique folder
	files := dblist.ReadFilesFromPaths(uniquedirs)

	keepfiles := make(map[string][]dblist.FileInfoWin)

	// Append to dontDeleteLastbackupfiles all files that should NOT be deleted:
	// the most recent files and files that have no config lines for them.
	for dir, filesindir := range files {
		// get the slice of most recent files. This files will not be deleted.
		dontDeleteLastbackupfiles := dblist.GetLastFilesGroupedByFunc(filesindir, dblist.GroupFunc, uniquesuffixes, keepLastNcopies)

		// A directory may contain some extra files not covered by config file - don't delete them.
		notInConfigFile := dblist.GetFilesNotCoveredByConfigFile(filesindir, conf, dblist.GroupFunc, uniquesuffixes)
		dontDeleteLastbackupfiles = append(dontDeleteLastbackupfiles, notInConfigFile...)

		// descending order expected in deleteArchivedFiles while searching in the list of last files.
		sort.Slice(dontDeleteLastbackupfiles, func(i, j int) bool {
			return dontDeleteLastbackupfiles[i].Name() > dontDeleteLastbackupfiles[j].Name() //DESC
		})
		keepfiles[dir] = dontDeleteLastbackupfiles // keepfiles needs to be ordered
	}
	deleteArchivedFiles(files, keepfiles, delArchived, dryRun)

}

// deleteArchivedFiles needs exceptfiles to be sorted descending.
// Exceptfiles sorted descending because it was used to get most recent backup file name earlier in the code.
func deleteArchivedFiles(files, exceptfiles map[string][]dblist.FileInfoWin, delArchived, dryrun bool) {
	for dir, slice := range files { // for every dir
		for _, finf := range slice { // for every file in dir
			if !finf.IsDir() { // dont touch subdirs
				// No ARCHIVE atrribute means that file has been archived or copied somewhere.
				// If file has Archive attribute you need a flag delArchived == true to delete this file.
				if (finf.WinAttr&syscall.FILE_ATTRIBUTE_ARCHIVE) == 0 || // no archive attribute set
					delArchived { // or you insist on deleting files with Archive attr set

					// search if current file is in the exception list
					keepfiles := exceptfiles[dir]                         // exceptfiles is already sorted descending
					pos := sort.Search(len(keepfiles), func(i int) bool { // we use bin search
						return keepfiles[i].Name() <= finf.Name() // <= for the descending slice

					})
					if pos < len(keepfiles) && keepfiles[pos].Name() == finf.Name() {
						// found in exceptfiles
						continue // file should not be deleted. this is the last file in the group of backup files
					}
					fullFilename, errpath := filepath.Abs(filepath.Join(dir, finf.Name()))

					log.Printf("rm %s\r\n", fullFilename)
					if !dryrun && errpath == nil {
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
