package main

import (
	"github.com/adamcolton/adasync/collection"
	"github.com/adamcolton/err"
	"os"
	"path/filepath"
	"time"
)

type Runable interface {
	Run()
}

type FullScan struct{}

func (_ FullScan) Run() {
	err.Debug("-- Starting Full Scan --")
	collection.FullScan()
}

type QuickScan struct{}

func (_ QuickScan) Run() {
	err.Debug("-- Checking for New Drives --")
	collection.QuickScan()
}

type SyncAll struct{}

func (_ SyncAll) Run() {
	err.Debug("-- Running Sync --")
	collection.SyncAll()
}

func main() {
	out, _ := os.Create("log.txt")
	err.DebugOut = out
	err.Debug("Started")
	path, e := filepath.Abs("./")
	err.Panic(e)
	path = filepath.ToSlash(path)
	settings, _ := collection.LoadConfig(path + "config.txt")
	if ignore, ok := settings["ignore"]; ok {
		collection.Settings["ignore"] = ignore
		err.Debug("Ignoring: ", ignore)
	} else {
		collection.Settings["ignore"] = collection.DefaultIgnore
		err.Debug("No ignore setting - using defaults")
	}
	runChan := make(chan Runable, 100)
	go func(runChan <-chan Runable) {
		for {
			run := <-runChan
			run.Run()
		}
	}(runChan)

	go func() {
		for {
			time.Sleep(time.Minute * 5)
			runChan <- QuickScan{}
			runChan <- SyncAll{}
		}
	}()

	for {
		runChan <- FullScan{}
		runChan <- SyncAll{}
		time.Sleep(time.Hour * 2)
	}
}
