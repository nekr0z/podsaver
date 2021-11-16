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
// along with this program. If not, see <https://www.gnu.org/licenses/>.

//go:generate go run version_generate.go

package main

import (
	"flag"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"io/ioutil"
	"os"
)

type episode struct {
	filename string // local filename
}

type podcast struct {
	local afero.Fs         // local filesystem for downloaded episodes
	url   string           // feed URL
	ep    map[int]*episode // key would be the episode number
}

var (
	output  io.Writer
	version string = "custom-build"
)

func init() {
	output = ioutil.Discard
}

func main() {
	output = os.Stdout

	fmt.Fprintf(output, "podsaver version %s\n", version)

	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	path := flag.String("p", wd, "location of the locally downloaded episodes")
	rename := flag.Bool("r", false, "rename the already downloaded episodes if needed")
	flag.Parse()
	var url string
	switch flag.NArg() {
	case 0:
		fmt.Fprintf(os.Stderr, "no feed URL is provided\n")
		os.Exit(1)
	case 1:
		url = flag.Arg(0)
	default:
		fmt.Fprintf(os.Stderr, "too many arguments\n")
		os.Exit(1)
	}

	pod := podcast{
		local: afero.NewBasePathFs(afero.NewOsFs(), *path),
		url:   url,
	}

	if err := pod.scanDir(); err != nil {
		fmt.Fprintf(os.Stderr, "error while scanning local episodes: %s\n", err)
		os.Exit(2)
	}
	fmt.Fprintf(output, "have %d episodes locally, last downloaded is #%d\n", len(pod.ep), pod.mostCurrent())

	if err := pod.matchFeed(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(3)
	}

	if *rename {
		if err := pod.renameDownloaded(); err != nil {
			fmt.Fprintf(os.Stderr, "error while renaming local files: %s\n", err)
		}
	}

	fmt.Fprintf(output, "all done!\n")
}
