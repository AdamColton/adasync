package collection

import (
	"fmt"
	"github.com/adamcolton/err"
	"path/filepath"
	"testing"
)

func TestMD5(t *testing.T) {
	testPathStr, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)

	file1 := &Path{
		root:   testPathStr,
		relDir: "/",
		name:   "md5test.txt",
	}
	file2 := &Path{
		root:   testPathStr,
		relDir: "/",
		name:   "md5test2.txt",
	}

	hash1, isDir, _ := file1.Stat()
	if len(hash1) != 16 {
		t.Error("Wrong hash length")
	}
	if isDir {
		t.Error("Is not dir")
	}

	hash2, _, _ := file2.Stat()
	expected := []byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}
	for i, b := range expected {
		if hash1[i] != b {
			t.Error("Hash1 is incorrect")
			break
		}
		if hash2[i] != b {
			t.Error("Hash2 is incorrect")
			break
		}
	}
}

func TestPathFromString(t *testing.T) {
	err.DebugEnabled = true
	a := PathFromString("C:\\testing\\foo.txt", "C:/testing")
	if a.root != "C:/testing" || a.relDir != "/" || a.name != "foo.txt" {
		t.Error("Fail 1")
		t.Error(a.root)
		t.Error(a.relDir)
		t.Error(a.name)
	}
	a = PathFromString("C:\\testing\\bar\\", "C:/testing")
	if a.root != "C:/testing" || a.relDir != "/" || a.name != "bar/" {
		t.Error("Fail 2")
		t.Error(a.root)
		t.Error(a.relDir)
		t.Error(a.name)
	}
	err.DebugEnabled = false
}

func TestPathFromNode(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = toSlash(testPath)
	ins := c.AddInstance(testPath)

	hash := Hash([...]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10})
	unchanged := ins.AddResource(&hash, 0, ins.root, "unchanged.txt")
	path := unchanged.PathNodes.nodes[0].RelativePath()
	if path.relDir != "/" {
		t.Error("Incorrect relative directory")
		fmt.Println("r: ", path.relDir)
	}
	if path.name != "unchanged.txt" {
		t.Error("Incorrect name path")
		fmt.Println("n: ", path.name)
	}

	hash = Hash([...]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	dir := ins.AddDirectory(&hash, ins.root, "foo/")
	path = dir.PathNodes.nodes[0].RelativePath()
	if path.relDir != "/" {
		t.Error("Incorrect relative directory")
		fmt.Println("r: ", path.relDir)
	}
	if path.name != "foo/" {
		t.Error("Incorrect name path")
		fmt.Println("n: ", path.name)
	}

}
