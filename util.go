package main

import (
	"bytes"
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

func compare(fs1 afero.Fs, file1 string, fs2 afero.Fs, file2 string) (bool, error) {
	f1, err := afero.ReadFile(fs1, file1)
	if err != nil {
		return false, err
	}
	f2, err := afero.ReadFile(fs2, file2)
	if err != nil {
		return false, err
	}
	return bytes.Equal(f1, f2), nil
}
