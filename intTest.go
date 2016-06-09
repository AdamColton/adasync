package main

//integration testing

import (
	"fmt"
	"github.com/adamcolton/adasync/adasync"
	"time"
)

type Runable interface {
	Run()
}

type FullScan struct{}

func (_ FullScan) Run() {
	fmt.Println("-- Starting Full Scan --")
	collection.FullScan()
}

type QuickScan struct{}

func (_ QuickScan) Run() {
	fmt.Println("-- Checking for New Drives --")
	collection.QuickScan()
}

type NullOp struct{}

func (_ NullOp) Run() {}

type SyncAll struct{}

func (_ SyncAll) Run() {
	fmt.Println("-- Running Sync --")
	collection.SyncAll()
}

func main() {
	collection.Settings["ignore"] = "C:/$,C:/gopath/src/github.com/adamcolton,D:/$,C:/Windows,D:/Music,D:/Video,D:/SteamLibrary"
	runChan := make(chan Runable, 10)
	watchdog := true
	go func(runChan <-chan Runable) {
		for {
			run := <-runChan
			run.Run()
			watchdog = true
		}
	}(runChan)
	start := time.Now()
	runChan <- FullScan{}
	runChan <- SyncAll{}
	runChan <- NullOp{}

	for {
		if len(runChan) == 0 {
			break
		}
		time.Sleep(time.Second)
	}
	go func() {
		fmt.Println("Starting watchdog")
		for {
			if !watchdog {
				break
			}
			watchdog = false
			time.Sleep(time.Second * 10)
		}
		panic("Watchdog timed out")
	}()
	d := time.Since(start)
	fmt.Println("Initial Scan complete: ", d.Seconds(), "seconds")
	fmt.Println("Sync will run every 15 seconds")

	for {
		for i := 0; i < 3; i++ {
			time.Sleep(time.Second * 5)
			watchdog = true
		}
		//runChan <- QuickScan{}
		runChan <- FullScan{}
		runChan <- SyncAll{}
	}
}
