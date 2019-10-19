package main

import (
	"flag"
	"fmt"
	"github.com/spf13/afero"
	"os"
)

type episode struct {
	filename string // local filename
	id       string // ID in feed
}

type podcast struct {
	local afero.Fs         // local filesystem for downloaded episodes
	url   string           // feed URL
	ep    map[int]*episode // key would be the episode number
}

func main() {
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
	fmt.Printf("have %d episodes locally, last downloaded is #%d\n", len(pod.ep), pod.mostCurrent())

	if err := pod.matchFeed(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(3)
	}

	if *rename {
		if err := pod.renameDownloaded(); err != nil {
			fmt.Fprintf(os.Stderr, "error while renaming local files: %s\n", err)
		}
	}

	fmt.Println("all done!")
}
