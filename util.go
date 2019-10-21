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
	"github.com/spf13/afero"
)

func copyFile(from, to afero.Fs, fromName, toName string) error {
	data, err := afero.ReadFile(from, fromName)
	if err != nil {
		return err
	}
	if err := afero.WriteFile(to, toName, data, 0644); err != nil {
		return err
	}
	return nil
}

func (pod *podcast) mostCurrent() (maxNumber int) {
	const maxUint = ^uint(0)
	const maxInt = int(maxUint >> 1)
	const minInt = -maxInt - 1
	maxNumber = minInt
	for n := range pod.ep {
		if n > maxNumber {
			maxNumber = n
		}
	}
	return maxNumber
}
