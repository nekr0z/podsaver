// +build integration

package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAerostat(t *testing.T) {
	fs := afero.NewBasePathFs(afero.NewOsFs(), "testdata/")

	httpFs := afero.NewHttpFs(fs)
	fileserver := http.FileServer(httpFs.Dir("/server"))
	server := httptest.NewServer(fileserver)

	pod := podcast{
		local: afero.NewBasePathFs(fs, "downloaded"),
		url:   fmt.Sprintf("%s/aerostat.rss", server.URL),
	}
	if err := pod.scanDir(); err != nil {
		t.Errorf("could not scan: %s", err)
	}

	if err := pod.matchFeed(); err != nil {
		t.Errorf("failed to match feed: %s", err)
	}

	if !deepCompare(t, fs, "downloaded/750.mp3", "golden/750.mp3") {
		t.Errorf("files for episode 750 do not match")
	}

	if err := fs.Remove("downloaded/750.mp3"); err != nil {
		t.Errorf("could not remove file: %s", err)
	}

	server.Close()
}

const chunkSize = 64000

func deepCompare(t *testing.T, fs afero.Fs, file1, file2 string) bool {
	f1, err := fs.Open(file1)
	if err != nil {
		t.Fatal(err)
	}
	defer f1.Close()

	f2, err := fs.Open(file2)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				t.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}
