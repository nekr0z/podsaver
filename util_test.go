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
	"testing"
)

func TestCopyFileErrors(t *testing.T) {
	fs := afero.NewMemMapFs()

	if err := copyFile(fs, fs, "foo", "bar"); err == nil {
		t.Fatalf("should have returned error for non-existing file")
	}

	if err := afero.WriteFile(fs, "foo", nil, 0644); err != nil {
		t.Fatal(err)
	}

	fsr := afero.NewReadOnlyFs(fs)

	if err := copyFile(fs, fsr, "foo", "bar"); err == nil {
		t.Fatalf("should have returned error for writing on read-only file")
	}
}

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
