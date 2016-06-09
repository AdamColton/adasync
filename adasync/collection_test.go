package adasync

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New()
	if len(c.id) != 16 {
		t.Error("Expected an ID of length 16")
	}
}

func TestIdStr(t *testing.T) {
	c := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)
	if c.IdStr() != "AQIDBAUGBwgJCgsMDQ4PEA==" {
		t.Error("Wrong ID string for collection")
	}
}

func TestAddInstance(t *testing.T) {
	filesystemA()
	c := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)
	pathStr := "/test/"
	ins := c.AddInstance(pathStr)
	if ins.pathStr != pathStr {
		t.Error("Incorrect path string")
	}
}
