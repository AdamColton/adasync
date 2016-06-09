// +build windows

package adasync

import (
	"fmt"
	"testing"
	"time"
)

func TestScanDrives(t *testing.T) {
	t.Skip() //annoying test requiring human interaction
	drives = make([]string, 0)
	detectDrives()
	fmt.Println(drives)
	fmt.Println("Plug in drive")
	time.Sleep(15 * time.Second)
	newDrives, _ := detectDrives()
	if len(newDrives) != 1 {
		t.Error("No new drive detected")
	}

	fmt.Println(newDrives, drives)
	fmt.Println("Remove drive")
	time.Sleep(15 * time.Second)
	_, missingDrives := detectDrives()
	if len(missingDrives) != 1 {
		t.Error("No drive missing")
	}
}

func TestFullAndQuick(t *testing.T) {
	t.Skip() //annoying test requiring human interaction
	drives = make([]string, 0)
	fmt.Println(fullScan())
	fmt.Println("Plug in drive")
	time.Sleep(15 * time.Second)
	fmt.Println(quickScan())
}
