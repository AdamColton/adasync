package collection

import (
	"io"
	"os"
)

var fs fileSystem = osFS{}

type fileSystem interface {
	Open(name string) (file, error)
	Stat(name string) (os.FileInfo, error)
	Create(name string) (file, error)
	Mkdir(name string, perm os.FileMode) error
	RemoveAll(path string) error
	Rename(oldpath, newpath string) error
}

type file interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	Stat() (os.FileInfo, error)
	Sync() error
}

// osFS implements fileSystem using the local disk.
type osFS struct{}

type FileInfo os.FileInfo

func (osFS) Open(name string) (file, error)            { return os.Open(name) }
func (osFS) Stat(name string) (os.FileInfo, error)     { return os.Stat(name) }
func (osFS) Create(name string) (file, error)          { return os.Create(name) }
func (osFS) Mkdir(name string, perm os.FileMode) error { return os.Mkdir(name, perm) }
func (osFS) RemoveAll(path string) error               { return os.RemoveAll(path) }
func (osFS) Rename(oldpath, newpath string) error      { return os.Rename(oldpath, newpath) }
