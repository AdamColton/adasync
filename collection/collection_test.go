package collection

import (
	"bytes"
	"fmt"
	"github.com/adamcolton/err"
	"os"
	"path/filepath"
	"strings"
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
	if len(d.added) != 5 {
		t.Error("Expected 5 files, got:")
		t.Error(len(d.added))
		for _, a := range d.added {
			t.Error(a)
		}
	}
}

func TestSelfDiffAllThree(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = toSlash(testPath)
	ins := c.AddInstance(testPath)

	unchanged := ins.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), 0, ins.root, "md5test.txt")
	moved := ins.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), 0, ins.root, "Moved.foo")
	deleted := ins.AddResource(HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17}), 0, ins.root, "deleted.bar")
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

}

func TestSerialized(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = toSlash(testPath)
	ins1 := c.AddInstance(testPath)
	ins1.SelfUpdate()
	b := ins1.Marshal()
	ins2 := Unmarshal(b, testPath+"2") // can't use actual path or it will return a ref
	if !bytes.Equal(ins2.collection.id, ins1.collection.id) {
		t.Error("Collection IDs do not match")
		fmt.Println(ins2.collection.id)
		fmt.Println(ins1.collection.id)
	}

	if md5test, ok := ins2.resources["rbHFRphzUsms0DGyOENzmg=="]; ok {
		if !strings.HasSuffix(md5test.FullPath(), "/testCollectionA2/subdir/test.txt") {
			t.Error("Bad path for subdir/test.txt")
			t.Error(md5test.FullPath())
		}
	} else {
		t.Error("Did not find subdir/test.txt")
		for hash, res := range ins2.resources {
			t.Error(hash, res.PathNodes.Last().Name)
		}
	}

	ins2.BadInstanceScan()

}

func TestWrite(t *testing.T) {
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	if _, err := os.Stat(testPath + "/.tag.collection"); !os.IsNotExist(err) {
		os.Remove(testPath + "/.tag.collection")
	}
	c := New()
	testPath = toSlash(testPath)
	ins := c.AddInstance(testPath)
	ins.SelfUpdate()
	ins.Write()
	if _, err := os.Stat(testPath + "/.tag.collection"); !os.IsNotExist(err) {
		t.Error(".tag.collection file should not exist in root")
		t.Error(testPath)
		os.Remove(testPath + "/.tag.collection")
	}
}

func TestRead(t *testing.T) {
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = toSlash(testPath)
	colFile, e := os.Open(testPath + "/.collection")
	defer colFile.Close()
	err.Test(e, t)
	b := make([]byte, 1024)
	l, e := colFile.Read(b)
	err.Test(e, t)
	ins := Unmarshal(b[:l], testPath)
	if len(ins.resources) != 4 {
		t.Error("Wrong number of resources")
		t.Error(len(ins.resources))
		for _, i := range ins.resources {
			t.Error(i.FullPath())
		}
	}
	if len(ins.directories) != 2 {
		t.Error("Wrong number of directories")
		t.Error(len(ins.directories))
	}
}

func TestOpen(t *testing.T) {
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPath = toSlash(testPath)
	ins := Open(testPath)
	if len(ins.resources) != 4 {
		t.Error("Wrong number of resources.")
		t.Error(len(ins.resources))
		for _, i := range ins.resources {
			t.Error(i.FullPath())
		}
	}
}

func TestOpenOnDNE(t *testing.T) {
	testPath, e := filepath.Abs("./")
	err.Test(e, t)
	testPath = toSlash(testPath)
	ins := Open(testPath)
	if len(ins.resources) != 0 {
		t.Error("Wrong number of resources.")
	}
}
