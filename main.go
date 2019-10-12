package main

import (
	"github.com/spf13/afero"
)

type episode struct {
	filename string
}

type podcast struct {
	local afero.Fs
	ep    map[int]*episode // key would be the episode number
}
