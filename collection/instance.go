package collection

import (
	"crypto/md5"
	"encoding/base64"
	"github.com/adamcolton/err"
	proto "github.com/golang/protobuf/proto"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const readOnlyIdLen = 10
const version = uint32(1)

var instances = make(map[Path]*Instance)

type Instance struct {
	collection  *Collection
	pathStr     string
	resources   map[string]*Resource
	directories map[string]*Directory
	root        *Directory
	settings    map[string]string
	dirty       bool
}

func (ins *Instance) generateResourceId(hash *Hash, path *PathNode) *Hash {
	if readOnlyId, ok := ins.settings["readonly"]; ok && len(readOnlyId) == readOnlyIdLen {
		out := append(hash[:], []byte(readOnlyId)...)
		hashOut := Hash(md5.Sum(out))
		err.Debug(hashOut)
		return &hashOut
	} else if ins.settings["AllowDuplicates"] == "false" {
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

func (ins *Instance) PathNode(parent *Directory, name string) *PathNode {
	pn := &PathNode{
		Name:     name,
		Instance: ins,
	}
	if parent != nil {
		pn.ParentID = parent.ID
	}
	return pn
}

func (ins *Instance) PathNodeFromHash(parentID *Hash, name string) *PathNode {
	return &PathNode{
		Name:     name,
		ParentID: parentID,
		Instance: ins,
	}
}

//both are used as sets, not maps
type deltaSelf struct {
	added         []string
	removed       map[string]*Resource
	removedByHash map[string]*Resource
	ins           *Instance
	deleted       map[string]*Resource
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
		ins:           ins,
		deleted:       make(map[string]*Resource),
	}
	// Put everything in removed and remove from removed
	// as we find each. What's left is what was actually
	// removed
	for _, res := range ins.resources {
		pn := res.PathNodes.Last()
		if pn.ParentID != nil || pn.Name != ".deleted" {
			fullPath := pn.FullPath()
			diff.removed[fullPath] = res
		} else if pn.Name == ".deleted" {
			diff.deleted[res.ID.String()] = res
		}
	}
	for _, dir := range ins.directories {
		pn := dir.PathNodes.Last()
		if pn.ParentID != nil || pn.Name != ".deleted" {
			fullPath := pn.FullPath()
			diff.removed[fullPath] = dir.Resource
		} else if pn.Name == ".deleted" {
			diff.deleted[dir.ID.String()] = dir.Resource
		}
	}
	filepath.Walk(ins.pathStr, diff.add)
	for _, res := range diff.removed {
		err.Debug("Removed: ", res.RelativePath().String())
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
	err.Debug("Self Update: ", ins.pathStr)
	diff := ins.SelfDiff()
	linkDirectories := make([]*Directory, 0)
	for _, newPathStr := range diff.added {
		newPath := PathFromString(newPathStr, ins.pathStr) //*Path
		pathNode := ins.PathToNode(newPath)                //*PathNode
		hash, isDir, size := newPath.Stat()
		if res, ok := diff.removedByHash[hash.String()]; ok {
			// resource was moved
			delete(diff.removed, res.FullPath())
			delete(diff.removedByHash, hash.String())
			res.PathNodes.Add(pathNode)
		} else {
			// resource is new
			if isDir {
				linkDirectories = append(linkDirectories, ins.AddDirectoryWithPath(hash, pathNode))
			} else {
				err.Debug("Added: ", newPathStr)
				ins.AddResourceWithPath(hash, size, pathNode)
			}
		}

		for _, dir := range linkDirectories {
			pathNode := dir.PathNodes.Last()
			ins.directories[pathNode.ParentID.String()].directories[pathNode.Name] = dir
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
	path := PathFromString(pathStr, d.ins.pathStr)
	if l := len(path.name); l >= 11 && strings.ToLower(path.name[l-11:]) == ".collection" {
		return nil
	}
	pathStr = path.String()
	if res, ok := d.removed[pathStr]; ok && d.ins.pathEqualsResource(res, pathStr) {
		delete(d.removed, pathStr)
	} else {
		err.Debug("Added: ", pathStr)
		d.added = append(d.added, pathStr)
	}
	return nil
}

func (ins *Instance) pathEqualsResource(res *Resource, pathStr string) bool {
	if checkLen, ok := ins.settings["check file length"]; ok && checkLen == "true" {
		if stat, e := os.Stat(pathStr); err.Warn(e) {
			pathSize := stat.Size()
			if stat.IsDir() {
				pathSize = 0
			}
			if pathSize != res.Size {
				err.Debug("SIZE DID NOT MATCH", stat.Size(), res.Size)
				return false
			}
		}
	}
	if checkHash, ok := ins.settings["check file hash"]; res != ins.root.Resource && ok && checkHash == "true" {
		hash, _, _ := PathFromString(pathStr, ins.pathStr).Stat()
		if hash != res.Hash {
			err.Debug("HASH DID NOT MATCH", hash, res.Hash)
			return false
		}
	}
	return true
}

func (ins *Instance) PathToNode(path *Path) *PathNode {
	cur := ins.root
	//the first will be "/", so we skip that
	paths := strings.SplitAfter(path.relDir, "/")[1:]
	if l := len(paths); l > 0 && paths[l-1] == "" {
		paths = paths[:l-1]
	}
	for _, dir := range paths {
		if dir == "" {
			continue
		}
		if next, ok := cur.directories[dir]; ok {
			cur = next
		} else {
			err.Debug(cur.directories)
			panic("Did not find " + dir)
		}
	}
	return &PathNode{
		Name:     path.name,
		ParentID: cur.ID,
		Instance: ins,
	}
}

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
	sIns, e = proto.Marshal(&VersionWrapper{
		Version:  version,
		Instance: sIns,
	})
	err.Warn(e)
	return sIns
}

func (sIns *SerialInstance) IdStr() string {
	return base64.StdEncoding.EncodeToString(sIns.CollectionId)
}

func Unmarshal(buf []byte, pathStr string) *Instance {
	versionWrapper := &VersionWrapper{}
	proto.Unmarshal(buf, versionWrapper)
	if versionWrapper.Version != version {
		// In the future, this will handle updating between versions
		// but right now there is only version 1
		panic("Incompatible Versions, check for update")
	}
	sIns := &SerialInstance{}
	proto.Unmarshal(versionWrapper.Instance, sIns)
	collection, ok := collections[sIns.IdStr()]
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
	for _, dir := range ins.directories {
		if parent := dir.PathNodes.Last().Parent(); parent != nil {
			parent.directories[dir.PathNodes.Last().Name] = dir
		}
	}
	for _, sRes := range sIns.Resources {
		sRes.unmarshalInto(ins)
	}
	return ins
}

func (ins *Instance) Write() {
	if !ins.dirty {
		return
	}
	if colFile, e := os.Create(ins.pathStr + "/.collection"); err.Log(e) {
		err.Debug("Writing: ", ins.pathStr)
		defer colFile.Close()
		colFile.Write(ins.Marshal())
		for _, dir := range ins.directories {
			if dir != ins.root {
				dir.WriteTag()
			}
		}
		ins.writeConfig()
		ins.dirty = false
	}
}

func (ins *Instance) writeConfig() {
	if configFile, e := os.Create(ins.pathStr + "/config.collection"); err.Log(e) {
		defer configFile.Close()
		ins.settings["id"] = ins.collection.IdStr()
		for key, val := range ins.settings {
			configFile.Write([]byte(key + ":" + val + "\n"))
		}
	}
}

func Open(pathStr string) *Instance {
	ins := loadInstance(pathStr)
	settings := LoadConfig(pathStr + "/config.collection")
	if ins == nil {
		var c *Collection
		if id, ok := settings["id"]; ok {
			if col, ok := collections[id]; ok {
				c = col
			} else {
				idBytes, e := base64.StdEncoding.DecodeString(id)
				err.Panic(e)
				c = New(idBytes...)
			}
		} else {
			c = New()
		}
		ins = c.AddInstance(pathStr)
		ins.dirty = true
	}
	ins.settings = settings
	if toLower(settings["readonly"]) == "true" {
		// if this is readonly, we need to give it a readonly ID
		readOnlyId := make([]rune, readOnlyIdLen)
		for i := 0; i < readOnlyIdLen; i++ {
			readOnlyId[i] = rune((rand.Float32() * 24) + 65)
		}
		settings["readonly"] = string(readOnlyId)
	}

	return ins
}

func loadInstance(pathStr string) *Instance {
	if colFile, e := os.Open(pathStr + "/.collection"); err.Check(e) {
		defer colFile.Close()
		if stat, e := colFile.Stat(); err.Log(e) {
			b := make([]byte, stat.Size())
			if l, e := colFile.Read(b); err.Log(e) {
				return Unmarshal(b[:l], pathStr)
			}
		}
	}
	return nil
}

// BadInstanceScan this is a debugging tool
// despite my best efforts, unit testing has not caught all the errors, this
// can help find additional errors under real conditions
func (ins *Instance) BadInstanceScan() {
	reset := err.DebugEnabled
	err.DebugEnabled = true

	for _, d := range ins.directories {
		pathNode := d.PathNodes.Last()
		name := pathNode.Name
		if name[len(name)-1] != '/' {
			err.Debug("--- Bad Directory Name", d.FullPath())
		}
		if name != ".deleted" {
			if root, ok := pathNode.getRoot(); !ok || root != ins.root.PathNodes.Last() {
				err.Debug("--- Bad root", d.FullPath())
			}
		}
		if pathNode.ParentID != nil {
			if _, ok := ins.directories[pathNode.ParentID.String()]; !ok {
				err.Debug("--- Did not find parent", d.FullPath())
			}
		}
	}

	for _, d := range ins.resources {
		pathNode := d.PathNodes.Last()
		name := pathNode.Name
		if name != ".deleted" {
			if root, ok := pathNode.getRoot(); !ok || root != ins.root.PathNodes.Last() {
				err.Debug("--- Bad root", d.FullPath())
			}
		}
		if pathNode.ParentID != nil {
			if _, ok := ins.directories[pathNode.ParentID.String()]; !ok {
				err.Debug("--- Did not find parent", d.FullPath(), pathNode.ParentID)
			}
		}
	}

	err.DebugEnabled = reset
}

func (p *PathNode) getRoot() (*PathNode, bool) {
	if p.ParentID == nil {
		return p, true
	}
	pnt := p.Parent()
	if pnt == nil {
		return nil, false
	}
	last := pnt.PathNodes.Last()
	if last == nil {
		return nil, false
	}
	return last.getRoot()
}
