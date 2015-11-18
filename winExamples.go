package main

import (
	"fmt"
	"github.com/adamcolton/err"
	"os"
	"syscall"
	"unsafe"
)

func main() {
	fmt.Println(scanDrives2())
}

func supressErrorDialogs() {
	var mod = syscall.NewLazyDLL("kernel32.dll")
	var proc = mod.NewProc("SetThreadErrorMode")
	proc.Call(0, 0, 0, 0)
	fmt.Println("Supressed")
}

func msgbox() {
	var mod = syscall.NewLazyDLL("user32.dll")
	var proc = mod.NewProc("MessageBoxW")
	var MB_YESNOCANCEL = 0x00000003

	ret, _, _ := proc.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Done Title"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("This test is Done."))),
		uintptr(MB_YESNOCANCEL))
	fmt.Printf("Return: %d\n", ret)
}

func scanDrives() []string {
	supressErrorDialogs()
	var mod = syscall.NewLazyDLL("kernel32.dll")
	var proc = mod.NewProc("GetDriveTypeW")

	drives := []string{}
	for letter := 'A'; letter <= 'Z'; letter++ {
		str := string(letter) + ":\\"
		strPtr := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(str)))
		ret, _, _ := proc.Call(strPtr, 0, 0, 0)
		if ret == 3 || ret == 2 {
			drives = append(drives, str)
		}
	}
	return drives
}

func listDrives() {
	var mod = syscall.NewLazyDLL("kernel32.dll")
	var proc = mod.NewProc("GetDriveTypeW")

	for letter := 'A'; letter <= 'Z'; letter++ {
		str := string(letter) + ":\\"
		strPtr := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(str)))
		ret, _, _ := proc.Call(strPtr, 0, 0, 0)
		if ret > 1 {
			fmt.Println(str, ret)
		}
	}
}

func scanDrives2() []string {
	wd, e := os.Getwd()
	err.Panic(e)
	drives := []string{}
	for letter := 'A'; letter <= 'Z'; letter++ {
		str := string(letter) + ":\\"
		e = os.Chdir(str)
		if e == nil {
			drives = append(drives, str)
		}
	}
	e = os.Chdir(wd)
	err.Panic(e)
	return drives
}
