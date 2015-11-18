package collection

import (
	"github.com/adamcolton/err"
	"os"
)

type Resource struct {
	ID        *Hash
	Hash      *Hash
	PathNodes *PathNodes
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
}

func (res *Resource) Serialize() *SerialResource {
	return &SerialResource{
		ID:        res.ID[:],
		Hash:      res.Hash[:],
		PathNodes: res.PathNodes.marshal(),
	}
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
		},
		directories: make(map[string]*Directory),
		resources:   make(map[string]*Resource),
	}
	ins.directories[dir.Hash.String()] = dir
	return dir
}

func (ins *Instance) AddResource(hash *Hash, parent *Directory, name string) *Resource {
	return ins.AddResourceWithPath(hash, ins.PathNode(parent, name))
}

func (ins *Instance) AddResourceWithPath(hash *Hash, pathNodes ...*PathNode) *Resource {
	res := &Resource{
		ID:        ins.collection.generateResourceId(hash, pathNodes[0]),
		Hash:      hash,
		PathNodes: NewPathNodes(0, pathNodes...),
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
	pns := NewPathNodes(0, pathNodes...)
	var id *Hash
	if tagFile, e := os.Open(pns.Last().FullPath() + ".tag.collection"); err.Check(e) {
		idBuf := make([]byte, 16)
		if l, e := tagFile.Read(idBuf); err.Log(e) && l == 16 {
			id = HashFromBytes(idBuf)
		}
	}
	if id == nil {
		id = ins.collection.generateResourceId(hash, pathNodes[0])
	}
	dir := &Directory{
		Resource: &Resource{
			ID:        id,
			Hash:      hash,
			PathNodes: pns,
		},
		directories: make(map[string]*Directory),
		resources:   make(map[string]*Resource),
	}
	ins.directories[dir.ID.String()] = dir
	return dir
}

func (dir *Directory) WriteTag() {
	tagFile, e := os.Create(dir.FullPath() + ".tag.collection")
	if err.Log(e) {
		defer tagFile.Close()
		tagFile.Write(dir.ID[:])
		//fmt.Println(dir.RelativePath().String(), dir.ID[:])
	}
}
