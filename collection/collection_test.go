package collection

import (
	"bytes"
	"fmt"
	"github.com/adamcolton/err"
	"os"
	"path/filepath"
	"testing"
)

var _ = fmt.Println

func TestNew(t *testing.T) {
	c := New()
	if len(c.id) != 16 {
		t.Error("Expected 16")
	}
}

func TestInstance(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)

	c.AddInstance(testPath)
}

func TestSelfDiffAdded(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)

	ins := c.AddInstance(testPath)
	d := ins.SelfDiff()
	if len(d.added) != 4 {
		t.Error("Expected 4 files, got:")
		t.Error(len(d.added))
	}
}

func TestSelfDiffAllThree(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = filepath.ToSlash(testPath)
	ins := c.AddInstance(testPath)

	unchanged := ins.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), ins.root, "md5test.txt")
	moved := ins.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), ins.root, "Moved.foo")
	deleted := ins.AddResource(HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17}), ins.root, "deleted.bar")
	ins.SelfUpdate()

	if len(unchanged.PathNodes.nodes) != 1 {
		t.Error("unchanged; expected 1, got:")
		t.Error(len(unchanged.PathNodes.nodes))
	}
	if len(moved.PathNodes.nodes) != 2 {
		t.Error("moved; expected 2, got:")
		t.Error(len(moved.PathNodes.nodes))
	} else if moved.PathNodes.Last().RelativePath().String() != "/md5test2.txt" {
		t.Error("moved; expected /md5test2.txt ; got: " + moved.PathNodes.Last().RelativePath().String())
	}
	if len(deleted.PathNodes.nodes) != 2 {
		t.Error("deleted; expected 2, got:")
		t.Error(len(deleted.PathNodes.nodes))
	}

	//fmt.Println(ins.root.directories["subdir"].FullPath())
}

func TestSerialized(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = filepath.ToSlash(testPath)
	ins := c.AddInstance(testPath)
	ins.SelfUpdate()
	b := ins.Marshal()
	sIns := Unmarshal(b)
	if !bytes.Equal(sIns.CollectionId, ins.collection.id) {
		t.Error("Collection IDs do not match")
		fmt.Println(sIns.CollectionId)
		fmt.Println(ins.collection.id)
	}
}

func TestWrite(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = filepath.ToSlash(testPath)
	ins := c.AddInstance(testPath)
	ins.SelfUpdate()
	ins.Write()
}

func TestRead(t *testing.T) {
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = filepath.ToSlash(testPath)
	colFile, e := os.Open(testPath + "/.collection")
	defer colFile.Close()
	err.Test(e, t)
	b := make([]byte, 1024)
	l, e := colFile.Read(b)
	err.Test(e, t)
	sIns := Unmarshal(b[:l])
	if len(sIns.Resources) != 3 {
		t.Error("Wrong number of resources")
		t.Error(len(sIns.Resources))
	}
	if len(sIns.Directories) != 1 {
		t.Error("Wrong number of directories")
		t.Error(len(sIns.Directories))
	}
}

func TestOpen(t *testing.T) {
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = filepath.ToSlash(testPath)
	ins := Open(testPath)
	if len(ins.resources) != 3 {
		t.Error("Wrong number of resources.")
		t.Error(len(ins.resources))
	}
}

func TestOpenOnDNE(t *testing.T) {
	testPath, e := filepath.Abs("./")
	err.Test(e, t)
	testPath = filepath.ToSlash(testPath)
	ins := Open(testPath)
	if len(ins.resources) != 0 {
		t.Error("Wrong number of resources.")
	}
}
