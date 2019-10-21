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
	"testing"
)

func TestMostCurrent(t *testing.T) {
	testCases := []struct {
		epis []int
		max  int
	}{
		{epis: []int{1, 2, 3, 4, 5, 6}, max: 6},
		{epis: []int{6}, max: 6},
	}

	for _, testCase := range testCases {
		pod := podcast{ep: make(map[int]*episode)}

		for _, i := range testCase.epis {
			pod.ep[i] = &episode{}
		}

		max := pod.mostCurrent()

		if max != testCase.max {
			t.Errorf("want %d, got %d", testCase.max, max)
		}
	}
}
