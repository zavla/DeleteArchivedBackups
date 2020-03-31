// package main leaves recent backups only. It doesn't delete last backup files.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	main()
}

func TestMainParamExample(t *testing.T) {
	printExample = true
	oldstdout := os.Stdout
	name := "./testdata/config.json"
	f, err := os.OpenFile(name, os.O_CREATE, 0)
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
