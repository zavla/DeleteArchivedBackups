// package main leaves recent backups only. It doesn't delete last backup files.

package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var skipdateinlog = func(s string, length int) string {
	pos := strings.LastIndexAny(s, "\\/")
	pos++
	return s[pos : pos+length]
}

func Test_main(t *testing.T) {
	printExample = false
	dryRun = true
	delArchived = true // to delete files with A attribute set
	configfile = "./testdata/config.json"
	logfile = "./testdata/logfile.tmp"
	os.Truncate(logfile, 0) // empty logfile for TESTs only

	_ = testcase_manyfiles(t)

	// as if we were run by scheduler
	main()

	// "want" - expects these files to be deleted
	want := []string{
		"зап_в_кам_2021-08-09T10-04-00-750-differ.rar",
		"dbase1_2021-08-01T17-37-00-360-FULL.rar",
	}

	CompareLines(t, logfile, want, 1, skipdateinlog)

}

func TestMain_badconfig(t *testing.T) {
	printExample = false
	dryRun = true
	delArchived = true // to delete files with A attribute set
	configfile = "./testdata/configBad.json"
	logfile = "./testdata/logfile.tmp"
	os.Truncate(logfile, 0) // empty logfile for TESTs only

	// as if we were run by scheduler
	main()

	// "want" - expects these files to be deleted
	want := []string{
		"Config file read error:",
	}
	extract := func(s string, length int) string {
		p := s[20 : 20+length]
		return p
	}
	CompareLines(t, logfile, want, 1, extract)

}

func TestMain_actualremove(t *testing.T) {
	printExample = false
	dryRun = false
	delArchived = true // to delete files with A attribute set
	configfile = "./testdata/configActualRemove.json"
	logfile = "./testdata/logfile.tmp"
	os.Truncate(logfile, 0)

	_ = createTestFiles(t, "./testdata/files/bases818", []string{
		"stange name",
		"dbase1_2021-08-01T12-12-12-FULL.rar",
		"dbase1_20210-222-3-FULL.rar",
		"dbase1_2021-08-01T13-12-12-FULL.rar",
		"dbase1_2021-07-01T13-12-12-FULL.rar",
	})

	main()

	want := []string{"dbase1_2021-07-01T13-12-12-FULL.rar"}

	CompareLines(t, logfile, want, 1, skipdateinlog)
}

// CompareLines compares lines from file with "want" slice of strings.
func CompareLines(t *testing.T, logfile string, want []string, startline int, extract func(string, int) string) {
	needlines := 1 << 32 // all lines
	got, err := Readlines(logfile, 1, needlines)
	if err != nil {
		t.Errorf("%v", err)
	}
	ok := true
	if len(got) != len(want) {
		ok = false
	}
	errline := 1
	for i, expectline := range want {
		errline = i + 1

		if i < len(got) {

			compare := extract(got[i], len(expectline))
			if expectline != compare {
				ok = false

				break
			}
			continue
		}
		break
	}
	if !ok {
		t.Errorf("testname %v:\n got %v\n want %v\n at line# %v", t.Name(), got, want, errline)
	}

}

func Readlines(filename string, startlinenumber int, countlines int) (ret []string, reterr error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(f)
	i := 1
	nread := 0
	for ; s.Scan(); i++ {
		if i < startlinenumber {
			continue
		}
		nread++
		ret = append(ret, s.Text())
		if nread >= countlines {
			break
		}
	}
	return ret, s.Err()
}

func testcase_manyfiles(t *testing.T) []string {
	bases116 := []string{
		"Monitor.7z",
		"ZUP_UDB_3_2021-03-24T11-09-31-793-FULL.bak",
		"ZUP_UDB_3_2021-08-02T14-36-41-130-FULL.rar",
		"ZUP_UDB_3_2021-08-06T17-40-00-440-FULL.rar",
		"beego.go",
		"buh_cap_DB_2021-08-01T17-40-00-330-FULL.rar",
		"buh_cap_DB_2021-08-06T17-40-00-430-FULL.rar",
		"buh_cap_DB_2021-08-09T10-03-00-700-differ.rar",
		"buh_cap_DB_2021-08-09T15-03-00-963-differ.rar",
		"buh_cap_DB_2021-08-10T10-03-00-703-differ.rar",
		"buh_cool7_2021-07-29T17-50-00-957-FULL.rar",
		"buh_cool7_2021-08-05T17-50-00-877-FULL.rar",
		"buh_log3_2021-08-01T17-40-00-450-FULL.rar",
		"buh_log3_2021-08-06T17-40-00-370-FULL.rar",
		"buh_log3_2021-08-09T10-03-00-720-differ.rar",
		"buh_log3_2021-08-09T15-03-00-977-differ.rar",
		"buh_log3_2021-08-10T10-03-00-710-differ.rar",
		"buh_log7_2021-07-29T17-20-00-463-FULL.rar",
		"buh_log7_2021-08-05T17-20-07-003-FULL.rar",
		"buh_log8_2021-08-01T17-40-00-363-FULL.rar",
		"buh_log8_2021-08-06T17-40-00-430-FULL.rar",
		"buh_log8_2021-08-09T10-01-03-750-differ.rar",
		"buh_log8_2021-08-09T15-01-02-440-differ.rar",
		"buh_log8_2021-08-10T10-01-01-327-differ.rar",
		"buh_pro7_2021-07-29T17-30-00-853-FULL.rar",
		"buh_pro7_2021-08-05T17-30-00-960-FULL.rar",
		"dbase1_2021-08-01T17-37-00-360-FULL.rar",
		"dbase1_2021-08-06T17-37-00-190-FULL.rar",
		"dbase1_2021-08-06T18-57-00-110-FULL.rar",
		"buh_pro8_2021-08-09T10-01-00-700-differ.rar",
		"buh_pro8_2021-08-09T15-01-06-093-differ.rar",
		"buh_pro8_2021-08-10T10-01-01-323-differ.rar",
		"ooo_UDB_distr_v3_2021-08-01T17-40-00-483-FULL.rar",
		"ooo_UDB_distr_v3_2021-08-06T17-40-00-963-FULL.rar",
		"sklad7_2021-07-29T05-30-32-170-FULL.rar",
		"sklad7_2021-08-05T05-30-03-130-FULL.rar",
		"srvinfo_192_168_11_6_1641_2021-05.rar",
		"srvinfo_192_168_11_6_1641_2021-07.rar",
		"zup_log8_2021-03-24T11-15-16-023-FULL.bak",
		"zup_log8_2021-08-02T14-39-59-873-FULL.rar",
		"zup_log8_2021-08-06T17-40-00-393-FULL.rar",
		"КипМ_2021-08-01T17-40-00-447-FULL.rar",
		"КипМ_2021-08-06T17-40-00-460-FULL.rar",
		"зап_в_кам_2021-08-01T17-47-03-337-FULL.rar",
		"зап_в_кам_2021-08-06T17-47-01-147-FULL.rar",
		"зап_в_кам_2021-08-09T10-04-00-750-differ.rar",
		"зап_в_кам_2021-08-09T15-04-01-063-differ.rar",
		"зап_в_кам_2021-08-10T10-04-00-717-differ.rar",
	}

	ret := createTestFiles(t, "./testdata/files/bases116", bases116)
	return ret
}

func createTestFiles(t *testing.T, dir string, files []string) (ret []string) {
	for _, filename := range files {

		fullname := filepath.Join(dir, filename)
		fullname, _ = filepath.Abs(fullname)
		_, err := os.Stat(fullname)
		if os.IsNotExist(err) {
			f, err := os.Create(fullname)
			if err != nil {
				t.Fatal(err)
			}
			f.Close()

		}
		ret = append(ret, fullname)
	}
	return ret
}

func TestMainParamExample(t *testing.T) {
	printExample = true
	oldstdout := os.Stdout
	name := "./testdata/configExample.json"
	f, err := os.OpenFile(name, os.O_CREATE, 0600)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	os.Stdout = f
	defer func() { os.Stdout = oldstdout }()

	main()

	os.Stdout = oldstdout
	_ = f.Close()
	f, err = os.Open(name)
	if err != nil {
		t.Error(err)
		return
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(b, []byte(exampleconf+"\n")) { // \n was added with fmt.Println()
		t.Errorf("not equal\n want %v\n got %v\n", []byte(exampleconf), b)
		return
	}
}
