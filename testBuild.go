package main

import (
	"github.com/adamcolton/adasync/adasync"
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
	cols := adasync.Scan([]string{
		thisPath,
	}).Slice()
	for _, pathStr := range cols {
		err.Debug("Found: ", pathStr)
		adasync.Open(pathStr)
	}
	adasync.SyncAll()
	err.Debug("Done")
}
