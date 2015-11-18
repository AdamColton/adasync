package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	walkList("./")
}

func walkList(path string) {
	path, _ = filepath.Abs(path)
	path = filepath.ToSlash(path)
	filepath.Walk(path, func(path string, _ os.FileInfo, _ error) error {
		fmt.Println(path)
		return nil
	})
}

/*
-- cannot use bytes as map keys --
b1 := []byte{3, 1, 4, 1, 5}
b2 := []byte{3, 1, 4, 1, 5}
mp := make(map[[]byte]string)
mp[b1] = "testing"
s, ok := mp[b2]
fmt.Println(s, ok)
*/
func byteSliceAsKey() {
	b1 := []byte{3, 1, 4, 1, 5}
	b2 := []byte{3, 1, 4, 1, 5}
	mp := make(map[string]string)
	mp[base64.StdEncoding.EncodeToString(b1)] = "testing"
	s, ok := mp[base64.StdEncoding.EncodeToString(b2)]
	fmt.Println(s, ok)
}

func splitExample() {
	path := "some\\directory\\following\\path.txt"
	dir, file := filepath.Split(path)
	fmt.Println("Dir: ", dir)
	fmt.Println("Name:", file)

	path = "some\\directory\\following\\"
	dir, file = filepath.Split(path)
	fmt.Println("Dir: ", dir)
	fmt.Println("Name:", file)
}

func pathToList() {
	path := "some\\directory\\following\\path.txt"
	path = filepath.ToSlash(path)
	pathList := strings.Split(path, "/")
	fmt.Println(pathList)
}

func waitForEnter() {
	fmt.Print("Plug in drive and hit enter")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func listFiles() {
	files, _ := ioutil.ReadDir("./")
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

/*
Notes:
 - It took 9.2s to scan 438,701 files on C: (more files, but ssd)
 - It took 14.4s to scan 92,621 files on D: (less files, but usb platter)
*/
func walkFilepath(path string) {
	c := 0
	start := time.Now()
	filepath.Walk(path, func(path string, _ os.FileInfo, _ error) error {
		c += 1
		return nil
	})
	d := time.Since(start)
	fmt.Println(d.Seconds(), c)
}

func matchFilename(drives []string) {
	for _, drive := range drives {
		filepath.Walk(drive, func(path string, fi os.FileInfo, _ error) error {
			dir, file := filepath.Split(path)
			if file == "config.collection" {
				fmt.Println(dir)
			}
			return nil
		})
	}
}

func md5Example() {
	data := []byte("These pretzels are making me thirsty.")
	hash := md5.Sum(data)
	fmt.Printf("%x\n", hash)
	fmt.Println(len(hash))
}

///---- This doesn't work----------
type foo []string

func (f *foo) Add(s string) {
	rf := *f
	rf = append(rf, s)
}

func tryAdd_fails() {
	f := &foo{}
	f.Add("test")
	fmt.Println(f)
}

// this does

type bar struct {
	f []string
}

func (b *bar) Add(s string) {
	b.f = append(b.f, s)
}

func tryAdd() {
	b := bar{}
	b.Add("test")
	fmt.Println(b)
}
