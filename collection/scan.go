package collection

import (
	"os"
	"path/filepath"
	"strings"
)

/*
Scan needs to find all the paths to collections
Then find the intersection and differences between
those paths and the one's were tracking
we don't do anything for the intersection
new paths need to be add
missing paths are dropped.
*/
func FullScan() []string {
	return scan(full()).Slice()
}

func QuickScan() []string {
	return scan(quick()).Slice()
}

type CollectionPaths struct {
	paths map[string]bool
}

func (cp *CollectionPaths) Slice() []string {
	s := make([]string, len(cp.paths))
	i := 0
	for k := range cp.paths {
		if l := len(k); k[l-1] == '/' {
			k = k[:l-1]
		}
		s[i] = k
		i++
	}
	return s
}

func scan(paths []string) *CollectionPaths {
	collectionPaths := &CollectionPaths{
		paths: make(map[string]bool),
	}
	for _, path := range paths {
		filepath.Walk(string(path), collectionPaths.AddIfCollection)
	}
	return collectionPaths
}

func (cp *CollectionPaths) AddIfCollection(subPath string, _ os.FileInfo, _ error) error {
	subPath = filepath.ToSlash(subPath)
	dir, file := filepath.Split(subPath)
	if l := len(file); l >= 11 && strings.ToLower(file[l-11:]) == ".collection" {
		cp.paths[dir] = true
		return filepath.SkipDir
	}
	return nil
}
