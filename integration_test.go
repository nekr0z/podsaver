// +build integration

package main

import (
	"fmt"
	"github.com/spf13/afero"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
)

func TestAerostat(t *testing.T) {
	fs := afero.NewBasePathFs(afero.NewOsFs(), "testdata/")

	httpFs := afero.NewHttpFs(fs)
	fileserver := http.FileServer(httpFs.Dir("/server"))
	server := httptest.NewServer(fileserver)

	cmd := exec.Command("./podsaver", "-r", "-p", "testdata/downloaded", fmt.Sprintf("%s/aerostat.rss", server.URL))
	cmd.Stdout = os.Stdout
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

	if err := fs.Rename("downloaded/015.mp3", "downloaded/15.mp3"); err != nil {
		t.Errorf("015.mp3 could not be renamed back: %s", err)
	}

	server.Close()
}
