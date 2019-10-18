package main

import (
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/spf13/afero"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestMatchFeed(t *testing.T) {
	pod, server := generatePodcast(t, 15, 10, 10, 2)
	if err := pod.scanDir(); err != nil {
		t.Fatal(err)
	}

	if err := pod.matchFeed(); err != nil {
		t.Fatal(err)
	}

	server.Close()

	for i := 2; i <= 15; i++ {
		if pod.ep[i] == nil {
			t.Fatalf("episode %d does not exist", i)
		}
		if pod.ep[i].filename == "" {
			t.Fatalf("episode %d has no local filename", i)
		}
	}
	for i := 12; i <= 15; i++ {
		if pod.ep[i].filename != strconv.Itoa(i)+".pcast" {
			t.Fatalf("filename for episode %d (%s) does not match the expected (%s)", i, pod.ep[i].filename, strconv.Itoa(i)+".pcast")
		}
	}
}

func TestMatchFeedIncomplete(t *testing.T) {
	pod, server := generatePodcast(t, 15, 10, 2, 11)
	if err := pod.scanDir(); err != nil {
		t.Fatal(err)
	}

	if err := pod.matchFeed(); err != nil {
		t.Fatal(err)
	}

	server.Close()

	for i := 11; i <= 15; i++ {
		if pod.ep[i] == nil {
			t.Fatalf("episode %d does not exist", i)
		}
		if pod.ep[i].filename == "" {
			t.Fatalf("episode %d has no local filename", i)
		}
	}
	for i := 13; i <= 15; i++ {
		if pod.ep[i].filename != strconv.Itoa(i)+".pcast" {
			t.Fatalf("filename for episode %d (%s) does not match the expected (%s)", i, pod.ep[i].filename, strconv.Itoa(i)+".pcast")
		}
	}
}

func TestMatchFeedErrors(t *testing.T) {
	pod := podcast{}
	if err := pod.matchFeed(); err != errNoFeed {
		t.Fatal("expected error for podcast with no feed")
	}

	pod, server := generatePodcast(t, 5, 2, 2, 1)
	if err := pod.matchFeed(); err != errNotScanned {
		t.Fatalf("expected error - not scanned local dir")
	}

	if err := pod.scanDir(); err != nil {
		t.Fatal(err)
	}

	if err := pod.matchFeed(); err != errNoMatch {
		t.Fatalf("expected error: can not possibly match; actual error: %s", err)
	}

	pod.url = "http://localhost/notthere"
	if err := pod.matchFeed(); err == nil {
		t.Fatal("expected error for unreachable feed, got nil")
	} else if _, ok := err.(feedParseError); !ok {
		t.Fatalf("expected error for unreachable feed, got: %s", err)
	}

	pod.local = nil
	if err := pod.matchFeed(); err == nil || err == errNoMatch {
		t.Fatalf("expected error for lack of local fs")
	}
	server.Close()
}

func generatePodcast(t *testing.T, episodes, feedSize, downloaded, firstDownloaded int) (podcast, *httptest.Server) {
	t.Helper()

	fs := afero.NewMemMapFs()
	for i := 1; i <= episodes; i++ {
		filename := fmt.Sprintf("episode%02d.pcast", i)
		generateRandomFile(t, fs, filename, 2048)
	}

	lfs := afero.NewMemMapFs()
	for i := firstDownloaded; i < downloaded+firstDownloaded; i++ {
		filename := fmt.Sprintf("episode%02d.pcast", i)
		if err := copyFile(fs, lfs, filename, filename); err != nil {
			t.Fatal(err)
		}
	}

	rfs := afero.NewMemMapFs()
	if err := rfs.Mkdir("/http", 0755); err != nil {
		t.Fatal(err)
	}

	httpFs := afero.NewHttpFs(rfs)
	fileserver := http.FileServer(httpFs.Dir("/http"))
	server := httptest.NewServer(fileserver)
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "mock podcast",
		Link:        &feeds.Link{Href: server.URL},
		Description: "test-test-test",
		Created:     now,
	}
	for i := 0; i < feedSize; i++ {
		filename := fmt.Sprintf("episode%02d.pcast", episodes-i)
		if err := copyFile(fs, rfs, filename, "/http/"+filename); err != nil {
			t.Fatal(err)
		}
		now = now.AddDate(0, 0, -14)
		feed.Add(&feeds.Item{
			Id:          strconv.Itoa(episodes - i),
			Link:        &feeds.Link{Href: server.URL},
			Title:       "yet another episode",
			Description: "latest and greatest",
			Enclosure: &feeds.Enclosure{
				Url:    fmt.Sprintf("%s/%s", server.URL, filename),
				Length: "1024",
				Type:   "media/podcast",
			},
			Created: now,
		})
	}
	rss, err := feed.ToRss()
	if err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(rfs, "/http/feed.rss", []byte(rss), 0644); err != nil {
		t.Fatal(err)
	}

	return podcast{
		local: lfs,
		url:   fmt.Sprintf("%s/feed.rss", server.URL),
	}, server
}

func generateRandomFile(t *testing.T, fs afero.Fs, name string, maxBytes int) {
	t.Helper()
	data := make([]byte, rand.Intn(maxBytes))
	if err := afero.WriteFile(fs, name, data, 0644); err != nil {
		t.Fatal(err)
	}
}
