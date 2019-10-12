package main

import (
	"github.com/spf13/afero"
	"reflect"
	"sort"
	"testing"
)

type ep struct {
	n    int
	name string
}

func TestScanDir(t *testing.T) {
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
		fs := populate(t, testCase.files)
		eps := scan(t, fs)
		assertEpisodes(t, eps, testCase.epis)
	}
}

func TestScanDirWithDirectories(t *testing.T) {
	files := []string{"100.avi", "101.avi", "102.mp3"}
	want := []ep{{100, "100.avi"}, {101, "101.avi"}, {102, "102.mp3"}}
	fs := populate(t, files)

	if err := fs.Mkdir("103", 0775); err != nil {
		t.Fatal(err)
	}

	eps := scan(t, fs)
	assertEpisodes(t, eps, want)
}

func TestScanDirError(t *testing.T) {
	pod := podcast{}

	if err := pod.scanDir(); err == nil {
		t.Error("scanned nothing")
	}

	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/bar/")

	pod.local = fs

	if err := pod.scanDir(); err == nil {
		t.Error("scanned unreadable directory")
	}
}

func populate(t *testing.T, filenames []string) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()

	for _, filename := range filenames {
		if err := afero.WriteFile(fs, filename, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	return fs
}

func scan(t *testing.T, fs afero.Fs) []ep {
	t.Helper()
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

	return eps
}

func assertEpisodes(t *testing.T, got, want []ep) {
	t.Helper()
	sort.Slice(got, func(i, j int) bool {
		return got[i].n < got[j].n
	})

	sort.Slice(want, func(i, j int) bool {
		return want[i].n < want[j].n
	})

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}
