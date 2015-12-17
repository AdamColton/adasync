// +build windows

package collection

import (
	"os"
)

/*
TODO: option to only scan new drives
*/

var DefaultIgnore = "C:/Program Files, C:/Program Files (x86), C:/Windows"

func osIgnore(path string) bool {
	return len(path) > 4 && path[1:4] == ":/$"
}

// full returns all drives
func full() []string {
	detectDrives()
	return drives
}

// quick returns only new drives that have appeared since last scan
func quick() []string {
	newDrives, _ := detectDrives()
	return newDrives
}

var drives = make([]string, 0)

func detectDrives() ([]string, []string) {
	drvs := []string{}
	newDrives := []string{}
	missingDrives := []string{}
	i := 0
	for letter := 'A'; letter <= 'Z'; letter++ {
		path := string(letter) + ":\\"
		if _, e := os.Open(path); e != nil {
			continue
		}
		drvs = append(drvs, path)
		for i < len(drives) && drives[i] < path {
			missingDrives = append(missingDrives, drives[i])
			i++
		}
		if i < len(drives) && drives[i] == path {
			i++
		} else {
			newDrives = append(newDrives, path)
		}
	}
	for ; i < len(drives); i++ {
		missingDrives = append(missingDrives, drives[i])
	}
	drives = drvs
	return newDrives, missingDrives
}
