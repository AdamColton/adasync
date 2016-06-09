package collection

import (
	"testing"
)

func TestGetSetting(t *testing.T) {
	filesystemA()
	c := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)
	pathStr := "/test/"
	ins := c.AddInstance(pathStr)
	if ins.GetSetting("static") != "true" {
		t.Error("Default static setting should be true")
	}
}
