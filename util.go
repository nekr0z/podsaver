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
