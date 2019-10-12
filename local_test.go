package main

import (
	"github.com/spf13/afero"
	"reflect"
	"sort"
	"testing"
)

func TestScanDir(t *testing.T) {
	type ep struct {
		n    int
		name string
	}

	var testCases = []struct {
		files []string
		epis  []ep
	}{
		{
			[]string{"1", "2", "3", "4", "5"},
			[]ep{
				{1, "1"},
				{2, "2"},
				{3, "3"},
				{4, "4"},
				{5, "5"},
			},
		},
		{
			[]string{"1.mp3", "2.mp3", "3.mp3", "5.mp3", "episode.7.mp3"},
			[]ep{
				{1, "1.mp3"},
				{2, "2.mp3"},
				{3, "3.mp3"},
				{5, "5.mp3"},
				{7, "episode.7.mp3"},
			},
		},
		{
			[]string{"01.mp3", "02.mp3", "25.mp3", "drop.txt", "26.mp3", "extras.mp3"},
			[]ep{
				{1, "01.mp3"},
				{2, "02.mp3"},
				{25, "25.mp3"},
				{26, "26.mp3"},
			},
		},
	}

	for _, testCase := range testCases {
		vfs := afero.NewMemMapFs()

		if err := vfs.MkdirAll("/foo/bar", 0755); err != nil {
			t.Fatal(err)
		}

		fs := afero.NewBasePathFs(vfs, "/foo/bar")

		for _, epi := range testCase.files {
			if err := afero.WriteFile(fs, epi, nil, 0644); err != nil {
				t.Fatal(err)
			}
		}

		pod := podcast{local: fs}

		if err := pod.scanDir(); err != nil {
			t.Fatal(err)
		}

		eps := make([]ep, len(pod.ep))

		i := 0
		for n, episode := range pod.ep {
			eps[i].n, eps[i].name = n, episode.filename
			i++
		}

		sort.Slice(eps[:], func(i, j int) bool {
			return eps[i].n < eps[j].n
		})

		sort.Slice(testCase.epis, func(i, j int) bool {
			return testCase.epis[i].n < testCase.epis[j].n
		})

		if !reflect.DeepEqual(eps, testCase.epis) {
			t.Errorf("want %v, got %v", testCase.epis, eps)
		}
	}
}
