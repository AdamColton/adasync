package collection

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/adamcolton/err"
	proto "github.com/golang/protobuf/proto"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var _ = fmt.Println

type Hash [md5.Size]byte

func HashFromBytes(bs []byte) *Hash {
	if len(bs) != 16 {
		panic("Requires 16 bytes")
	}
	hash := &Hash{}
	for i, b := range bs {
		hash[i] = b
	}
	return hash
}

func (hash *Hash) String() string {
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (a *Hash) Equal(b *Hash) bool {
	if a == nil && b == nil {
		return true
	}
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}
	return bytes.Equal(a[:], b[:])
}

type Collection struct {
	id        []byte
	instances map[string]*Instance
	settings  map[string]string
}

var collections = make(map[string]*Collection)

func (col *Collection) generateResourceId(hash *Hash, path *PathNode) *Hash {
	if col.settings["AllowDuplicates"] == "false" {
		return hash
	}
	out := append(hash[:], []byte(path.Name)...)
	parent := path.Parent()
	if parent != nil {
		out = append(out, parent.ID[:]...)
	}
	hashOut := Hash(md5.Sum(out))
	return &hashOut
}

func New(id ...byte) *Collection {
	if len(id) == 0 {
		id = make([]byte, 16)
		rand.Read(id)
	}
	c := &Collection{
		id:        id,
		instances: make(map[string]*Instance),
		settings:  make(map[string]string),
	}
	collections[base64.StdEncoding.EncodeToString(c.id)] = c
	return c
}

type Instance struct {
	collection  *Collection
	pathStr     string
	resources   map[string]*Resource
	directories map[string]*Directory
	root        *Directory
}

func (c *Collection) AddInstance(pathStr string) *Instance {
	pathStr = filepath.ToSlash(pathStr)
	ins := &Instance{
		collection:  c,
		pathStr:     pathStr,
		resources:   make(map[string]*Resource),
		directories: make(map[string]*Directory),
	}
	hash := &Hash{}
	ins.root = ins.AddDirectory(hash, nil, "/")
	c.instances[pathStr] = ins
	return ins
}

var instances = make(map[Path]*Instance)

//both are used as sets, not maps
type deltaSelf struct {
	added         []string
	removed       map[string]*Resource
	removedByHash map[string]*Resource
	instancePath  string
}

// SelfDiff
// Todo: handle directories
//
// A note on hashes: if there were two copies of a file and not there is only
// one and it has moved, we can't tell which one it was, the way we treat hashes
// will pick one
func (ins *Instance) SelfDiff() *deltaSelf {
	diff := &deltaSelf{
		removed:       make(map[string]*Resource),
		removedByHash: make(map[string]*Resource),
		instancePath:  ins.pathStr,
	}
	// Put everything in removed and remove from removed
	// as we find each. What's left is what was actually
	// removed
	for _, res := range ins.resources {
		fullPath := ins.pathStr + res.RelativePath().String()
		diff.removed[fullPath] = res
	}
	for _, dir := range ins.directories {
		fullPath := ins.pathStr + dir.RelativePath().String()
		diff.removed[fullPath] = dir.Resource
	}
	filepath.Walk(ins.pathStr, diff.add)
	for _, res := range diff.removed {
		diff.removedByHash[res.Hash.String()] = res
	}
	/**
		  Need to sort diff.added so that if we have a new file in a new folder, we
	    will add the folder before the file. This also puts the root folder first
	    so we can remove it
	*/
	sort.Sort(ByLength(diff.added))
	return diff
}

func (ins *Instance) SelfUpdate() {
	diff := ins.SelfDiff()
	for _, newPathStr := range diff.added {
		newPath := PathFromString(newPathStr, ins.pathStr)
		pathNode := ins.PathToNode(newPath)
		hash, isDir := newPath.Stat()
		if res, ok := diff.removedByHash[hash.String()]; ok {
			// resource was moved
			delete(diff.removed, ins.pathStr+res.RelativePath().String())
			delete(diff.removedByHash, hash.String())
			res.PathNodes.Add(pathNode)
		} else {
			// resource is new
			if isDir {
				ins.directories[pathNode.ParentID.String()].directories[pathNode.Name] = ins.AddDirectoryWithPath(hash, pathNode)
			} else {
				ins.AddResourceWithPath(hash, pathNode)
			}
		}

	}
	for _, res := range diff.removed {
		res.PathNodes.Add(ins.PathNodeFromHash(nil, ".deleted"))
	}
}

// add adds a path to a deltaSelf. If the path is in "removed"
// then it's a know resource and it's removed from removed
// if not, then it's a new resources and is added to added.
func (d *deltaSelf) add(pathStr string, fi os.FileInfo, _ error) error {
	if fi.IsDir() {
		pathStr += "/"
	}
	path := PathFromString(pathStr, d.instancePath)
	if l := len(path.name); l >= 11 && strings.ToLower(path.name[l-11:]) == ".collection" {
		return nil
	}
	pathStr = path.String()
	if _, ok := d.removed[pathStr]; ok {
		delete(d.removed, pathStr)
	} else {
		d.added = append(d.added, pathStr)
	}
	return nil
}

func (ins *Instance) PathToNode(path *Path) *PathNode {
	cur := ins.root
	//the first will be "/", so we skip that
	for _, dir := range strings.SplitAfter(path.relDir, "/")[1:] {
		if dir == "" {
			continue
		}
		if next, ok := cur.directories[dir]; ok {
			cur = next
		} else {
			fmt.Println(cur.directories)
			panic("Did not find " + dir)
		}
	}
	return &PathNode{
		Name:     path.name,
		ParentID: cur.ID,
		Instance: ins,
	}
}

/*
type Instance struct {
  collection  *Collection
  pathStr     string
  resources   []*Resource
  directories []*Directory
  root        *Directory
}
*/
func (ins *Instance) Marshal() []byte {
	sRes := make([]*SerialResource, len(ins.resources))
	sDirs := make([]*SerialResource, len(ins.directories)-1)
	i := 0
	for _, res := range ins.resources {
		sRes[i] = res.Serialize()
		i++
	}
	i = 0
	for _, dir := range ins.directories {
		if dir != ins.root {
			sDirs[i] = dir.Serialize()
			i++
		}
	}
	sIns, e := proto.Marshal(&SerialInstance{
		CollectionId: ins.collection.id[:],
		Resources:    sRes,
		Directories:  sDirs,
	})
	err.Warn(e)
	return sIns
}

func Unmarshal(buf []byte) *SerialInstance {
	sIns := &SerialInstance{}
	proto.Unmarshal(buf, sIns)
	return sIns
}

func (ins *Instance) Write() {
	colFile, e := os.Create(ins.pathStr + "/.collection")
	if err.Log(e) {
		defer colFile.Close()
		colFile.Write(ins.Marshal())
		for _, dir := range ins.directories {
			dir.WriteTag()
		}
	}
}

func Open(pathStr string) *Instance {
	// --- Start file stuff ---
	if colFile, e := os.Open(pathStr + "/.collection"); e == nil {
		defer colFile.Close()
		if stat, e := colFile.Stat(); err.Log(e) {
			b := make([]byte, stat.Size())
			if l, e := colFile.Read(b); err.Log(e) {
				// --- Get Collection ---
				sIns := Unmarshal(b[:l])
				collection, ok := collections[base64.StdEncoding.EncodeToString(sIns.CollectionId)]
				if !ok {
					collection = New(sIns.CollectionId...)
				} else if ins, ok := collection.instances[pathStr]; ok {
					// oh, we already had a handle to that instance
					return ins
				}
				// --- populate instance ---
				ins := collection.AddInstance(pathStr)
				for _, sDir := range sIns.Directories {
					sDir.unmarshalDirInto(ins)
				}
				for str, dir := range ins.directories {
					if parent := dir.PathNodes.Last().Parent(); parent != nil {
						parent.directories[str] = dir
					}
				}
				for _, sRes := range sIns.Resources {
					sRes.unmarshalInto(ins)
				}

				return ins
			}
		}
	} else if os.IsNotExist(e) {
		c := New()
		ins := c.AddInstance(pathStr)
		return ins
	} else {
		err.Log(e)
	}
	return nil
}
