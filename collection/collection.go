package collection

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

var _ = fmt.Println

var Settings = map[string]string{}

type Collection struct {
	id        []byte
	instances map[string]*Instance
}

var collections = make(map[string]*Collection)

func (col *Collection) IdStr() string {
	return base64.StdEncoding.EncodeToString(col.id)
}

func New(id ...byte) *Collection {
	if len(id) == 0 {
		id = make([]byte, 16)
		rand.Read(id)
	}
	c := &Collection{
		id:        id,
		instances: make(map[string]*Instance),
	}
	collections[c.IdStr()] = c
	return c
}

func (c *Collection) AddInstance(pathStr string) *Instance {
	pathStr = toSlash(pathStr)
	ins := &Instance{
		collection:  c,
		pathStr:     pathStr,
		resources:   make(map[string]*Resource),
		directories: make(map[string]*Directory),
		settings:    make(map[string]string),
	}
	hash := &Hash{}
	ins.root = ins.AddDirectory(hash, nil, "/")
	c.instances[pathStr] = ins
	return ins
}
