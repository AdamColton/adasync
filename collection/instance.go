package collection

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"github.com/adamcolton/err"
	proto "github.com/golang/protobuf/proto"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var instances = make(map[Path]*Instance)

type Instance struct {
	collection  *Collection
	pathStr     string
	resources   map[string]*Resource
	directories map[string]*Directory
	root        *Directory
	settings    map[string]string
}

func (ins *Instance) generateResourceId(hash *Hash, path *PathNode) *Hash {
	if ins.settings["AllowDuplicates"] == "false" {
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
		pn := res.PathNodes.Last()
		if pn.ParentID != nil || pn.Name != ".deleted" {
			fullPath := pn.FullPath()
			diff.removed[fullPath] = res
		}
	}
	for _, dir := range ins.directories {
		pn := dir.PathNodes.Last()
		if pn.ParentID != nil || pn.Name != ".deleted" {
			fullPath := pn.FullPath()
			diff.removed[fullPath] = dir.Resource
		}
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
			delete(diff.removed, res.FullPath())
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
	return sIns
}

func (sIns *SerialInstance) IdStr() string {
	return base64.StdEncoding.EncodeToString(sIns.CollectionId)
}

func Unmarshal(buf []byte) *SerialInstance {
	sIns := &SerialInstance{}
	proto.Unmarshal(buf, sIns)
	return sIns
}

func (ins *Instance) Write() {
	if colFile, e := os.Create(ins.pathStr + "/.collection"); err.Log(e) {
		defer colFile.Close()
		colFile.Write(ins.Marshal())
		for _, dir := range ins.directories {
			dir.WriteTag()
		}
		if configFile, e := os.Create(ins.pathStr + "/config.collection"); err.Log(e) {
			defer configFile.Close()
			configFile.Write([]byte("ID:" + ins.collection.IdStr()))
		}
	}
}

func Open(pathStr string) *Instance {
	ins := loadInstance(pathStr)
	settings := loadConfig(pathStr)

	if ins == nil {
		var c *Collection
		if id, ok := settings["id"]; ok {
			c = collections[id]
		} else {
			c = New()
		}
		ins = c.AddInstance(pathStr)
	}
	return ins
}

func loadConfig(pathStr string) map[string]string {
	settings := make(map[string]string)
	if configFile, e := os.Open(pathStr + "/config.collection"); err.Check(e) {
		defer configFile.Close()
		reader := bufio.NewReader(configFile)
		if line, e := reader.ReadBytes('\n'); err.Check(e) {
			setting := strings.SplitN(string(line), ":", 2)
			if len(setting) == 2 {
				settings[strings.ToLower(setting[0])] = setting[1]
			}
		}
	}
	return settings
}

func loadInstance(pathStr string) *Instance {
	if colFile, e := os.Open(pathStr + "/.collection"); err.Check(e) {
		defer colFile.Close()
		if stat, e := colFile.Stat(); err.Log(e) {
			b := make([]byte, stat.Size())
			if l, e := colFile.Read(b); err.Log(e) {
				// --- Get Collection ---
				sIns := Unmarshal(b[:l])
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
	}
	return nil
}
