package collection

import (
	"github.com/adamcolton/err"
	"path/filepath"
	"testing"
)

/*
Note: may have an issue with directory selfSync.
If a directory changes location, the ID is useless
we could add a file for tagging - probably the best short term answer

later, I can do a complex matching based on the directory contents

need to check skipDir on .collection
*/

func TestDiff(t *testing.T) {
	c := New()
	testPathA, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)
	testPathA = filepath.ToSlash(testPathA)
	insA := c.AddInstance(testPathA)

	insA.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), insA.root, "md5test.txt")
	insA.AddResource(HashFromBytes([]byte{80, 246, 243, 125, 31, 47, 211, 96, 69, 20, 35, 235, 227, 207, 10, 10}), insA.root, "Moved.foo")
	insA.AddResource(HashFromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17}), insA.root, "deleted.bar")
	insA.SelfUpdate()

	testPathB, e := filepath.Abs("../testCollectionB")
	err.Test(e, t)
	testPathB = filepath.ToSlash(testPathB)
	insB := c.AddInstance(testPathB)
	insB.SelfUpdate()
	sync := &Sync{
		a: insA,
		b: insB,
	}
	sync.Diff()
	flags := map[string]bool{
		".deleted":    false,
		"md5test.txt": false,
	}
	for _, action := range sync.actions {
		switch a := action.(type) {
		case *CpRes:
			flags[a.res.PathNodes.Last().Name] = true
		}
	}
	if !flags[".deleted"] {
		t.Error("Missed '.deleted'")
	}
	if !flags["md5test.txt"] {
		t.Error("Missed 'md5test.txt'")
	}
}
