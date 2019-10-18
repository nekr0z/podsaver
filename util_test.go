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
