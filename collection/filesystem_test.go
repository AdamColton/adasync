// this file does not contain any tests but include functions to mock the
// filesystem for testing
package collection

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

var EOF = errors.New("EOF")

type MockFileSystem struct {
	files map[string]*MockFile
}

func (mfs *MockFileSystem) dirExists(name string) error {
	dir, _ := filepath.Split(name)
	if _, ok := mfs.files[dir]; ok {
		return nil
	}
	return errors.New("open " + name + ": no such file or directory")
}

func (mfs *MockFileSystem) Open(name string) (file, error) {
	if f, ok := mfs.files[name]; ok {
		return f, nil
	}
	if e := mfs.dirExists(name); e != nil {
		return nil, e
	}
	return mfs.AddFile(name, ""), nil
}

func (mfs *MockFileSystem) Create(name string) (file, error) {
	if e := mfs.dirExists(name); e != nil {
		return nil, e
	}
	return mfs.AddFile(name, ""), nil
}

func (mfs *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	f, ok := mfs.files[name]
	if !ok {
		return nil, errors.New("stat " + name + ": no such file or directory")
	}
	return f.Stat()
}

func (mfs *MockFileSystem) Mkdir(name string, perm os.FileMode) error {
	dir, dirname := filepath.Split(name)
	if _, ok := mfs.files[dir]; !ok {
		return errors.New("mkdir " + name + ": no such file or directory")
	}
	mfs.files[name] = &MockFile{
		data: []byte{},
		pos:  0,
		name: dirname,
		info: MockFileInfo{
			name:    dirname,
			size:    0,
			mode:    perm,
			modTime: time.Now(),
			isDir:   true,
		},
	}
	return nil
}

func (mfs *MockFileSystem) RemoveAll(path string) error {
	panic("Not implemented")
	return errors.New("Not implemented")
}

func (mfs *MockFileSystem) Rename(oldpath, newpath string) error {
	panic("Not implemented")
	return errors.New("Not implemented")
}

func (mfs *MockFileSystem) AddFile(name, contents string) *MockFile {
	return mfs.AddBinaryFile(name, []byte(contents))
}

func (mfs *MockFileSystem) AddBinaryFile(name string, contents []byte) *MockFile {
	_, filename := filepath.Split(name)
	f := &MockFile{
		data: contents,
		pos:  0,
		name: filename,
		info: MockFileInfo{
			name:    filename,
			size:    0,
			mode:    0777,
			modTime: time.Now(),
			isDir:   false,
		},
	}
	mfs.files[name] = f
	return f
}

type MockFile struct {
	data []byte
	pos  int64
	name string
	info MockFileInfo
}

func (mf *MockFile) Close() error { return nil }
func (mf *MockFile) Sync() error  { return nil }

func (mf *MockFile) Stat() (os.FileInfo, error) {
	panic("Not implemented")
	return nil, errors.New("Not implemented")
}

func (mf *MockFile) Read(p []byte) (int, error) {
	n := copy(p, mf.data[mf.pos:])
	mf.pos += int64(n)
	e := error(nil)
	if mf.pos == int64(len(mf.data)) {
		e = EOF
	}
	return n, e
}

func (mf *MockFile) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, errors.New("offset cannot be negative")
	}
	if off >= int64(len(mf.data)) {
		return 0, EOF
	}
	n := copy(p, mf.data[off:])
	e := error(nil)
	if off+int64(n) == int64(len(mf.data)) {
		e = EOF
	}
	return n, e
}

func (mf *MockFile) Seek(offset int64, whence int) (int64, error) {
	l64 := int64(len(mf.data))
	switch whence {
	case 0:
		mf.pos = offset
	case 1:
		mf.pos += offset
	case 2:
		mf.pos = l64 + offset
	default:
		return 0, errors.New("Bad whence") //todo - get correct error
	}
	return mf.pos, nil
}

func (mf *MockFile) Write(p []byte) (n int, err error) {
	mf.data = append(mf.data[:mf.pos], p...)
	mf.pos += int64(len(p))
	return len(p), nil
}

type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (mfi MockFileInfo) Name() string       { return mfi.name }
func (mfi MockFileInfo) Size() int64        { return mfi.size }
func (mfi MockFileInfo) Mode() os.FileMode  { return mfi.mode }
func (mfi MockFileInfo) ModTime() time.Time { return mfi.modTime }
func (mfi MockFileInfo) IsDir() bool        { return mfi.isDir }
func (mfi MockFileInfo) Sys() interface{}   { return nil }

func filesystemA() {
	mfs := &MockFileSystem{
		files: make(map[string]*MockFile),
	}

	mfs.AddFile("config.txt", "id:1234\nfoo:bar")
	mfs.AddFile("/test/test.txt", "this is a test")

	fs = mfs

}
