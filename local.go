package main

import (
	"github.com/spf13/afero"
	"regexp"
	"strconv"
)

func (pod *podcast) scanDir() error {
	files, err := afero.ReadDir(pod.local, "/")
	if err != nil {
		return err
	}

	pod.ep = make(map[int]*episode)

	numRe := regexp.MustCompile(`\d+`)

	for _, file := range files {
		num, err := strconv.Atoi(numRe.FindString(file.Name()))
		if err != nil {
			continue
		}
		epi := episode{filename: file.Name()}
		pod.ep[num] = &epi
	}

	return nil
}
