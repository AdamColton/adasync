package collection

import (
	"fmt"
	"github.com/adamcolton/err"
)

var _ = fmt.Println

type Resource struct {
	ID        *Hash
	Hash      *Hash
	PathNodes *PathNodes
	Size      int64
}

func (r *Resource) RelativePath() *Path {
	return r.PathNodes.Last().RelativePath()
}

func (r *Resource) FullPath() string {
	return r.PathNodes.Last().FullPath()
}

type Directory struct {
	*Resource
	directories map[string]*Directory
	resources   map[string]*Resource
	tagged      bool
}

func (res *Resource) Serialize() *SerialResource {
	return &SerialResource{
		ID:        res.ID[:],
		Hash:      res.Hash[:],
		PathNodes: res.PathNodes.marshal(),
		Size:      res.Size,
	}
}

// Depth returns the directory depth of the resource
// This is used for syncing to make sure we sync all the folders in a directory
// before we start syncing their contents
func (res *Resource) Depth() int {
	i := 0
	pn := res.PathNodes.Last()
	for {
		parent := pn.Parent()
		if parent == nil {
			break
		}
		pn = parent.PathNodes.Last()
		i++
	}
	return i
}

func (sRes *SerialResource) unmarshalInto(ins *Instance) *Resource {
	pns := NewPathNodes(len(sRes.PathNodes))
	for i, spn := range sRes.PathNodes {
		pns.nodes[i] = spn.unmarshal(ins)
	}
	res := &Resource{
		ID:        HashFromBytes(sRes.ID),
		Hash:      HashFromBytes(sRes.Hash),
		PathNodes: pns,
		Size:      sRes.Size,
	}
	ins.resources[res.ID.String()] = res
	return res
}

func (sRes *SerialResource) unmarshalDirInto(ins *Instance) *Directory {
	pns := NewPathNodes(len(sRes.PathNodes))
	for i, spn := range sRes.PathNodes {
		pns.nodes[i] = spn.unmarshal(ins)
	}
	dir := &Directory{
		Resource: &Resource{
			ID:        HashFromBytes(sRes.ID),
			Hash:      HashFromBytes(sRes.Hash),
			PathNodes: pns,
			Size:      sRes.Size,
		},
		directories: make(map[string]*Directory),
		resources:   make(map[string]*Resource),
	}
	ins.directories[dir.ID.String()] = dir
	return dir
}

func (ins *Instance) AddResource(hash *Hash, size int64, parent *Directory, name string) *Resource {
	return ins.AddResourceWithPath(hash, size, ins.PathNode(parent, name))
}

func (ins *Instance) AddResourceWithPath(hash *Hash, size int64, pathNodes ...*PathNode) *Resource {
	resId := ins.generateResourceId(hash, pathNodes[0])
	if old, ok := ins.resources[resId.String()]; ok {
		//this (probably) means the resource was deleted and added again.
		old.Hash = hash
		old.Size = size
		old.PathNodes.nodes = append(old.PathNodes.nodes, pathNodes...)
		err.Debug(old.PathNodes.Last().FullPath())
		return old
	}
	res := &Resource{
		ID:        resId,
		Hash:      hash,
		PathNodes: NewPathNodes(0, pathNodes...),
		Size:      size,
	}
	ins.resources[res.ID.String()] = res
	return res
}

func (ins *Instance) AddDirectory(hash *Hash, parent *Directory, name string) *Directory {
	return ins.AddDirectoryWithPath(hash, ins.PathNode(parent, name))
}

func (ins *Instance) AddDirectoryWithPath(hash *Hash, pathNodes ...*PathNode) *Directory {
	if len(pathNodes) == 0 {
		panic("Cannot create Directory without path")
	}
	tagged := false
	pns := NewPathNodes(0, pathNodes...)
	var id *Hash
	pnsLast := pns.Last()
	if l := len(pnsLast.Name); l == 0 || pnsLast.Name[l-1] != '/' {
		panic("Bad directory name: " + pnsLast.Name)
	}
	if tagFile, e := fs.Open(pnsLast.FullPath() + ".tag.collection"); err.Check(e) {
		defer tagFile.Close()
		idBuf := make([]byte, 16)
		if l, e := tagFile.Read(idBuf); err.Log(e) && l == 16 {
			id = HashFromBytes(idBuf)
			tagged = true
		}
	}
	if id == nil {
		id = ins.generateResourceId(hash, pathNodes[0])
	}
  if old, ok := ins.directories[id.String()]; ok {
    //this (probably) means the directory was deleted and added again.
    old.tagged = tagged
    old.PathNodes.nodes = append(old.PathNodes.nodes, pathNodes...)
    err.Debug(old.PathNodes.Last().FullPath())
    return old
  }
	dir := &Directory{
		Resource: &Resource{
			ID:        id,
			Hash:      hash,
			PathNodes: pns,
		},
		tagged:      tagged,
		directories: make(map[string]*Directory), //maps the name to the directory, not the ID
		resources:   make(map[string]*Resource),
	}
	ins.directories[dir.ID.String()] = dir
	if pnsLast.ParentID == nil {
		if pnsLast.Name != "/" && pnsLast.Name != ".deleted" {
			panic("Bad Node: " + pnsLast.Name)
		}
	} else if parent, ok := ins.directories[pnsLast.ParentID.String()]; ok {
		parent.directories[pnsLast.Name] = dir
	}
	return dir
}

func (dir *Directory) WriteTag() {
	if pn := dir.PathNodes.Last(); !dir.tagged && !pn.IsDeleted() {
		tagFile, e := fs.Create(dir.FullPath() + ".tag.collection")
		if err.Log(e) {
			defer tagFile.Close()
			tagFile.Write(dir.ID[:])
			dir.tagged = true
		}
	}
}
