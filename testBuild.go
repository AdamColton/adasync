package main

import (
	"./collection"
	"github.com/adamcolton/err"
	"path/filepath"
)

/*
Use this for integration testing.
It will scan all subdirectories and sync them.
*/

func main() {
	err.DebugOut = err.Stdout
	thisPath, _ := filepath.Abs(".")
	cols := collection.Scan([]string{
		thisPath,
	}).Slice()
	for _, pathStr := range cols {
		err.Debug("Found: ", pathStr)
		collection.Open(pathStr)
	}
	collection.SyncAll()
	err.Debug("Done")
}
