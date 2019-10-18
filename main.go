package main

import (
	"github.com/spf13/afero"
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
