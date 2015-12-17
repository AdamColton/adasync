package collection

import (
	"fmt"
	"github.com/adamcolton/err"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var _ = fmt.Println

/*
Note: may have an issue with directory selfSync.
If a directory changes location, the ID is useless
we could add a file for tagging - probably the best short term answer

later, I can do a complex matching based on the directory contents

need to check skipDir on .collection
*/

func TestDiff(t *testing.T) {
	// copy the dirs
	testPathA, e := filepath.Abs("../syncTestCollectionA")
	err.Test(e, t)
	testPathB, e := filepath.Abs("../syncTestCollectionB")
	err.Test(e, t)
	sourceA, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	sourceB, e := filepath.Abs("../testCollectionB")
	err.Test(e, t)
	err.Test(os.RemoveAll(testPathA), t)
	err.Test(os.RemoveAll(testPathB), t)
	filepath.Walk(string(sourceA), func(sourceStr string, fi os.FileInfo, _ error) error {
		dstStr := testPathA + strings.Replace(sourceStr, sourceA, "", -1)
		if !fi.IsDir() {
			dstFile, _ := os.Create(dstStr)
			srcFile, _ := os.Open(sourceStr)
			defer dstFile.Close()
			defer srcFile.Close()
			io.Copy(dstFile, srcFile)
		} else {
			os.Mkdir(dstStr, 0700)
		}
		return nil
	})
	filepath.Walk(string(sourceB), func(sourceStr string, fi os.FileInfo, _ error) error {
		dstStr := testPathB + strings.Replace(sourceStr, sourceB, "", -1)
		if !fi.IsDir() {
			dstFile, _ := os.Create(dstStr)
			srcFile, _ := os.Open(sourceStr)
			defer dstFile.Close()
			defer srcFile.Close()
			io.Copy(dstFile, srcFile)
		} else {
			os.Mkdir(dstStr, 0700)
		}
		return nil
	})

	//start the test
	c := New()
	testPathA = toSlash(testPathA)
	insA := c.AddInstance(testPathA)

	insA.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), 0, insA.root, "md5test.txt")
	insA.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), 0, insA.root, "Moved.foo")
	insA.AddResource(HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17}), 0, insA.root, "deleted.bar")
	insA.SelfUpdate()

	testPathB = toSlash(testPathB)
	insB := c.AddInstance(testPathB)
	insB.SelfUpdate()
	sync := &Sync{
		a: insA,
		b: insB,
	}
	sync.Diff()
	flags := map[string]string{}
	for _, actionList := range sync.actions {
		for _, action := range actionList {
			switch a := action.(type) {
			case *CpRes:
				flags[a.res.PathNodes.Last().Name] = "CpRes"
			case *MvRes:
				str := a.cloneFrom.PathNodes.Last().Name + " -> " + a.cloneTo.PathNodes.Last().Name
				flags[str] = "MvRes"
			case *CpDir:
				fmt.Println("FOO", a.dir.RelativePath().String())
			default:
				t.Error("Missed One")
			}
		}
	}
	if flags[".deleted"] != "CpRes" {
		t.Error("Missed '.deleted'")
	}
	if flags["md5test.txt"] != "CpRes" {
		t.Error("Missed 'md5test.txt'")
	}
	if flags["md5test2.txt -> Moved.foo"] != "MvRes" {
		t.Error("Missed 'md5test2.txt -> Moved.foo'")
	}

	sync.Run()

	insA.BadInstanceScan()
	insB.BadInstanceScan()
}
