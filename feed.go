package main

import (
	"bytes"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var (
	errNoFeed  = errors.New("no feed available")
	errNoMatch = errors.New("could not match items in feed to local files")
)

type feedParseError struct {
	url string
	err error
}

func (p feedParseError) Error() string {
	return fmt.Sprintf("failed to read and parse feed from %s", p.url)
}

func (p feedParseError) Cause() error {
	return p.err
}

func (pod *podcast) matchFeed() error {
	if pod.url == "" {
		return errNoFeed
	}

	if err := pod.scanDir(); err != nil {
		return err
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(pod.url)
	if err != nil {
		return feedParseError{
			url: pod.url,
			err: err,
		}
	}

	sort.Slice(feed.Items, func(i, j int) bool {
		return feed.Items[i].PublishedParsed.After(*feed.Items[j].PublishedParsed)
	})
	lastDownloaded := pod.mostCurrent()

	pod.tmp = afero.NewMemMapFs()
	remote := make(map[int]string)
	var match int
	for i, item := range feed.Items {
		remote[i], err = downloadEpisode(pod.tmp, item)
		if err != nil {
			return err
		}
		if pod.compareFile(lastDownloaded, remote[i]) {
			pod.ep[i].id = item.GUID
		}
		if pod.ep[i] != nil && pod.ep[i].id != "" {
			match = i
			break
		}
	}
	if match == 0 {
		return errNoMatch
	}
	for i, n := match-1, lastDownloaded+1; i >= 0; i, n = i-1, n+1 {
		if feed.Items[i].GUID != "" {
			pod.ep[n] = &episode{id: feed.Items[i].GUID}
			filename := strconv.Itoa(n) + filepath.Ext(remote[i])
			if err := copyFile(pod.tmp, pod.local, remote[i], filename); err != nil {
				return err
			} else {
				pod.ep[n].filename = filename
			}
		}
	}

	return nil
}

func downloadEpisode(fs afero.Fs, item *gofeed.Item) (string, error) {
	if len(item.Enclosures) != 1 {
		return "", fmt.Errorf("no enclosure or multiple enclosures for episode")
	}
	if item.Enclosures[0].URL == "" {
		return "", fmt.Errorf("no remote URL for episode")
	}
	filename := fmt.Sprintf("%s.", item.GUID)
	name, err := downloadFile(fs, filename, item.Enclosures[0].URL)
	if err != nil {
		return "", err
	}
	return name, nil
}

func downloadFile(fs afero.Fs, filenamePrefix, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filename, err := guessFilename(resp)
	if err != nil {
		return "", err
	}

	filename = filenamePrefix + filename

	out, err := fs.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return filename, err
}

func guessFilename(resp *http.Response) (string, error) {
	filename := resp.Request.URL.Path
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			filename = params["filename"]
		}
	}

	// sanitize
	if filename == "" || strings.HasSuffix(filename, "/") || strings.Contains(filename, "\x00") {
		return "", fmt.Errorf("no filename")
	}

	filename = filepath.Base(filepath.Clean("/" + filename))
	if filename == "" || filename == "." || filename == "/" {
		return "", fmt.Errorf("no filename")
	}

	return filename, nil
}

func (pod *podcast) compareFile(n int, file string) bool {
	f1, err := afero.ReadFile(pod.local, pod.ep[n].filename)
	if err != nil {
		return false
	}
	f2, err := afero.ReadFile(pod.tmp, file)
	if err != nil {
		return false
	}
	return bytes.Equal(f1, f2)
}
