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
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	errNoEpisodes = errors.New("no episodes have been found")
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
		if file.IsDir() {
			continue
		}

		num, err := strconv.Atoi(numRe.FindString(strings.TrimSuffix(file.Name(), path.Ext(file.Name()))))
		if err != nil {
			continue
		}

		epi := episode{filename: file.Name()}
		pod.ep[num] = &epi
	}

	if len(pod.ep) == 0 {
		return errNoEpisodes
	}

	return nil
}

func (pod *podcast) renameDownloaded() error {
	max := pod.mostCurrent()
	d := len(strconv.Itoa(max))
	for n, ep := range pod.ep {
		ext := path.Ext(ep.filename)
		newname := fmt.Sprintf("%0*d%s", d, n, ext)
		if ep.filename != newname {
			if err := pod.local.Rename(ep.filename, newname); err != nil {
				return err
			}
		}
		ep.filename = newname
	}

	return nil
}
