package collection

import (
	"github.com/adamcolton/err"
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
func fullScan() []string {
	return Scan(full()).Slice()
}

func FullScan() {
	fs := fullScan()
	for _, pathStr := range fs {
		err.Debug("Found: ", pathStr)
		Open(pathStr)
	}
}

func quickScan() []string {
	return Scan(quick()).Slice()
}

func QuickScan() {
	for _, pathStr := range quickScan() {
		err.Debug("Found new drive: ", pathStr)
		Open(pathStr)
	}
}

func SyncAll() {
	for _, c := range collections {
		inss := make([]*Instance, len(c.instances))
		idx := 0
		for _, ins := range c.instances {
			inss[idx] = ins
			idx++
		}
		for i, ins := range inss {
			ins.SelfUpdate()
			if ins.dirty || ins.isNew {
				for j, prev := range inss {
					if j == i {
						break
					}
					sync := Sync{
						a:       ins,
						b:       prev,
						actions: make(map[int][]Action),
					}
					err.Debug("Syncing: ", ins.pathStr)
					err.Debug("     To: ", prev.pathStr)
					sync.Diff()
					if !sync.Run() {
						break
					}
				}
			}
			ins.isNew = false
		}
	}
	for _, c := range collections {
		for _, ins := range c.instances {
			ins.Write()
		}
	}
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

func Scan(paths []string) *CollectionPaths {
	collectionPaths := &CollectionPaths{
		paths: make(map[string]bool),
	}
	for _, path := range paths {
		filepath.Walk(string(path), collectionPaths.AddIfCollection)
	}
	return collectionPaths
}

func (cp *CollectionPaths) AddIfCollection(subPath string, _ os.FileInfo, _ error) error {
	subPath = toSlash(subPath)
	if ignLstStr, ok := Settings["ignore"]; ok {
		if osIgnore(subPath) {
			return filepath.SkipDir
		}
		ignLst := StringList(ignLstStr)
		for _, ignore := range ignLst {
			if len(ignore) > 0 && strings.HasPrefix(subPath, strings.Trim(ignore, " \n")) {
				err.Debug("Ignoring: ", subPath)
				return filepath.SkipDir
			}
		}
	}
	dir, file := filepath.Split(subPath)
	file = strings.ToLower(file)
	if file == "config.collection" || file == ".collection" {
		cp.paths[dir] = true
		return filepath.SkipDir
	}
	return nil
}
