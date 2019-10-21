// Copyright (C) 2019 Evgeny Kuznetsov (evgeny@kuznetsov.md)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along tihe this program. If not, see <https://www.gnu.org/licenses/>.

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
	errNoFeed      = errors.New("no feed available")
	errNoMatch     = errors.New("could not match items in feed to local files")
	errNoEnclosure = errors.New("no enclosure or multiple enclosures for episode")
	errNotScanned  = errors.New("no data on downloaded episodes - not scanned?")
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

	if pod.ep == nil {
		return errNotScanned
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

	match := -1
	for i, item := range feed.Items {
		if len(item.Enclosures) != 1 {
			return errNoEnclosure
		}
		if item.Enclosures[0].URL == "" {
			return fmt.Errorf("no remote URL for episode")
		}

		ok, err := compareRemote(item.Enclosures[0].URL, pod.local, pod.ep[lastDownloaded].filename)
		if err != nil {
			return err
		}
		if ok {
			match = i
			break
		}
	}
	if match == -1 {
		return errNoMatch
	}

	fmt.Fprintf(output, "Successfully matched feed to local directory content; episode #%d is #%d in the feed.\n", lastDownloaded, match)

	for i, n := match-1, lastDownloaded+1; i >= 0; i, n = i-1, n+1 {
		if len(feed.Items[i].Enclosures) != 1 {
			return errNoEnclosure
		}
		if err := pod.downloadEpisode(n, feed.Items[i].Enclosures[0].URL); err != nil {
			return err
		}
	}

	for i, n := 0, pod.mostCurrent(); i < len(feed.Items); i, n = i+1, n-1 {
		if _, ok := pod.ep[n]; !ok {
			if len(feed.Items[i].Enclosures) != 1 {
				return errNoEnclosure
			}
			if err := pod.downloadEpisode(n, feed.Items[i].Enclosures[0].URL); err != nil {
				return err
			}
		}
	}

	return nil
}

func (pod *podcast) downloadEpisode(n int, url string) error {
	fmt.Fprintf(output, "Fetching episode #%d: ", n)

	file, err := downloadFile(pod.local, strconv.Itoa(n), url)
	if err != nil {
		return err
	}
	pod.ep[n] = &episode{filename: file}
	return nil
}

func downloadFile(fs afero.Fs, filenamePrefix, url string) (string, error) {
	fmt.Fprintf(output, "downloading %s...", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filename, err := guessFilename(resp)
	if err != nil {
		return "", err
	}

	filename = filenamePrefix + filepath.Ext(filename)

	out, err := fs.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	fmt.Fprintf(output, "successfully downloaded %s\n", filename)

	return filename, err
}

func compareRemote(url string, fs afero.Fs, filename string) (bool, error) {
	const chunkSize = 1024 * 64

	fmt.Fprintf(output, "comparing %s to %s...\n", url, filename)
	info, err := fs.Stat(filename)
	if err != nil {
		return false, err
	}

	length := info.Size()

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	f, err := fs.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	bf := make([]byte, chunkSize)
	br := make([]byte, chunkSize)

	for i := 0; int64(i) < length/chunkSize; i++ {
		if _, err := io.ReadFull(f, bf); err != nil {
			return false, err
		}

		if _, err := io.ReadFull(resp.Body, br); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return false, nil
			}
			return false, err
		}

		if !bytes.Equal(bf, br) {
			return false, nil
		}
	}

	bf = make([]byte, length%chunkSize)
	br = make([]byte, length%chunkSize)

	if _, err := io.ReadFull(f, bf); err != nil {
		return false, err
	}

	if _, err := io.ReadFull(resp.Body, br); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return false, nil
		}
		return false, err
	}

	if !bytes.Equal(bf, br) {
		return false, nil
	}

	if n, err := resp.Body.Read(br); n != 0 || err != io.EOF {
		return false, nil
	}

	return true, nil
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
