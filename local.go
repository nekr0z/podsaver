package main

import (
	"fmt"
	"github.com/spf13/afero"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func (pod *podcast) scanDir() error {
	if pod.local == nil {
		return fmt.Errorf("no local location specified")
	}

	files, err := afero.ReadDir(pod.local, "/")
	if err != nil {
		return err
	}

	pod.ep = make(map[int]*episode)

	numRe := regexp.MustCompile(`\d+`)

	for _, file := range files {
		num, err := strconv.Atoi(numRe.FindString(strings.TrimSuffix(file.Name(), path.Ext(file.Name()))))
		if err != nil {
			continue
		}
		epi := episode{filename: file.Name()}
		pod.ep[num] = &epi
	}

	return nil
}
