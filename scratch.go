package main

import (
	"./collection"
	"github.com/adamcolton/err"
)

func main() {
	err.DebugEnabled = true
	collection.Open("C:/Users/Adam/Documents/adasync/test2")
	collection.Open("D:/adasync/test")
	collection.Open("G:/adasync/test3")
	collection.SyncAll()
	err.Debug("Done")
}
