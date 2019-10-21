// +build integration

package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestAerostat(t *testing.T) {
	fs := afero.NewBasePathFs(afero.NewOsFs(), "testdata/")

	httpFs := afero.NewHttpFs(fs)
	fileserver := http.FileServer(httpFs.Dir("/server"))
	server := httptest.NewServer(fileserver)

	cmd := exec.Command("./podsaver", "-r", "-p", "testdata/downloaded", fmt.Sprintf("%s/aerostat.rss", server.URL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		t.Errorf("podsaver returned error: %s", err)
	}

	ok, err := compare(fs, "downloaded/750.mp3", fs, "golden/750.mp3")
	if err != nil {
		t.Errorf("error comparing files: %s", err)
	}
	if !ok {
		t.Errorf("files for episode 750 do not match")
	}

	if err := fs.Remove("downloaded/750.mp3"); err != nil {
		t.Errorf("could not remove file: %s", err)
	}

	for i := 742; i <= 748; i++ {
		if err := fs.Remove("downloaded/" + strconv.Itoa(i) + ".mp3"); err != nil {
			t.Errorf("could not remove file: %s", err)
		}
	}

	if err := fs.Rename("downloaded/015.mp3", "downloaded/15.mp3"); err != nil {
		t.Errorf("015.mp3 could not be renamed back: %s", err)
	}

	server.Close()
}

func compare(fs1 afero.Fs, file1 string, fs2 afero.Fs, file2 string) (bool, error) {
	f1, err := afero.ReadFile(fs1, file1)
	if err != nil {
		return false, err
	}
	f2, err := afero.ReadFile(fs2, file2)
	if err != nil {
		return false, err
	}
	return bytes.Equal(f1, f2), nil
}
