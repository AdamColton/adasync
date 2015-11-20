package collection

import (
	"fmt"
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
	fs := Scan([]string{
		"C:/Users/Adam/Documents/runesync",
		"D:/runesync",
	}).Slice()

	//fs := fullScan()
	for _, pathStr := range fs {
		Open(pathStr)
	}
}

func quickScan() []string {
	return Scan(quick()).Slice()
}

func QuickScan() {
	for _, pathStr := range quickScan() {
		fmt.Println("Found new drive: ", pathStr)
		Open(pathStr)
	}
}

func SyncAll() {
	for _, c := range collections {
		var prev *Instance
		for _, ins := range c.instances {
			ins.SelfUpdate()
			if prev != nil {
				sync := Sync{
					a:       ins,
					b:       prev,
					actions: make([]Action, 0),
				}
				sync.Diff()
				sync.Run()
			}
			prev = ins
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
	subPath = filepath.ToSlash(subPath)
	if ignLstStr, ok := Settings["ignore"]; ok {
		ignLst := strings.Split(ignLstStr, ",")
		for _, ignore := range ignLst {
			if strings.HasPrefix(subPath, strings.Trim(ignore, " ")) {
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
