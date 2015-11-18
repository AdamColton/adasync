package collection

import (
	"github.com/adamcolton/err"
	"path/filepath"
	"testing"
)

func TestRelativePath(t *testing.T) {
	c := New()
	testPath, e := filepath.Abs("../testCollectionA")
	err.Test(e, t)

	ins := c.AddInstance(testPath)
	if ins.root.RelativePath().String() != "/" {
		t.Error("Root should be / got:")
		t.Error(ins.root.RelativePath().String())
	}
}
