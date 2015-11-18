package collection

import (
	"github.com/adamcolton/err"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	runesyncRoot, e := filepath.Abs("../")
	err.Test(e, t)
	collections := scan([]string{runesyncRoot})
	if len(collections.paths) != 2 {
		t.Error("Should have two collections")
		t.Error(collections.paths)
	}

	colPathStr := collections.Slice()[0]
	if l := len(colPathStr); l > 0 && colPathStr[l-1] == '/' {
		t.Error("Root path should not end with a slash")
	}
}
