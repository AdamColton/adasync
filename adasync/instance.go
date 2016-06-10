package adasync

import (
	"crypto/md5"
	"encoding/base64"
	"github.com/adamcolton/err"
	proto "github.com/golang/protobuf/proto"
	"math/rand"
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
	isNew       bool
}

// generateResourceId takes a resource hash and path and will generate an ID
// the process depends on instance settings.
// If the instance is read-only, the read-only value is appended to the hash
// and a new hash is generated
// If the instance does not allow duplicates
func (ins *Instance) generateResourceId(hash *Hash, path *PathNode) *Hash {
	if readOnlyId := ins.GetSetting("read only"); len(readOnlyId) == readOnlyIdLen {
		out := append(hash[:], []byte(readOnlyId)...)
		id := Hash(md5.Sum(out))
		err.Debug(id)
		return &id
	} else if ins.GetSetting("allow duplicates") == "false" {
		return hash
	}
	out := append(hash[:], []byte(path.Name)...)
	parent := path.Parent()
	if parent != nil {
		out = append(out, parent.ID[:]...)
	}
	id := Hash(md5.Sum(out))
	for {
		if _, idExists := ins.directories[id.String()]; idExists {
			id = Hash(md5.Sum(id.Bytes()))
			continue
		}
		if _, idExists := ins.resources[id.String()]; idExists {
			id = Hash(md5.Sum(id.Bytes()))
			continue
		}
		break
	}
	return &id
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
		err.Debug("Version) Expected: ", version, " Got:", versionWrapper.Version)
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
	if colFile, e := filesystem.Create(ins.pathStr + "/.collection"); err.Log(e) {
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
	if configFile, e := filesystem.Create(ins.pathStr + "/config.collection"); err.Log(e) {
		defer configFile.Close()
		ins.settings["id"] = ins.collection.IdStr()
		// write out all the default settings values
		for key, _ := range instanceDefaults {
			ins.settings[key] = ins.GetSetting(key)
		}
		for key, val := range ins.settings {
			configFile.Write([]byte(key + ":" + val + "\n"))
		}
	}
}

func Open(pathStr string) *Instance {
	ins := loadInstance(pathStr)
	settings, _ := LoadConfig(pathStr + "/config.collection")
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
	if toLower(settings["read only"]) == "true" {
		// if this is readonly, we need to give it a read only ID
		readOnlyId := make([]rune, readOnlyIdLen)
		for i := 0; i < readOnlyIdLen; i++ {
			readOnlyId[i] = rune((rand.Float32() * 24) + 65)
		}
		settings["read only"] = string(readOnlyId)
	}

	return ins
}

func loadInstance(pathStr string) *Instance {
	if colFile, e := filesystem.Open(pathStr + "/.collection"); err.Check(e) {
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

var instanceDefaults = map[string]string{
	"static":           "true",
	"read only":        "false",
	"allow duplicates": "true",
}

// GetSetting will return the setting for the instance. If the instance does
// not have a value for a standard setting, it will return the default value
func (ins *Instance) GetSetting(key string) string {
	if val, ok := ins.settings[key]; ok {
		return val
	}
	if val, ok := instanceDefaults[key]; ok {
		return val
	}
	return ""
}

// pathEqualsResource checks that a path and a resource at that path are
// considered equal. With no special configuration, we assume they are
// if either check hash or check length is true, we check that the file length
// is the same. If the file length is different we return false (potentially
// skipping the hash check). If check hash is true, we check the hash.
func (ins *Instance) pathEqualsResource(res *Resource, pathStr string) bool {
	checkHash := false
	if ins.GetSetting("static") == "false" {
		checkHash = true
	}
	if checkLen := ins.GetSetting("check file length"); (checkLen == "true") || checkHash {
		if stat, e := filesystem.Stat(pathStr); err.Warn(e) {
			pathSize := stat.Size()
			if stat.IsDir() {
				pathSize = 0
			}
			if pathSize != res.Size {
				err.Debug("Size did not match", stat.Size(), res.Size, pathStr)
				return false
			}
		}
	}
	if res != ins.root.Resource && checkHash {
		hash, _, _ := PathFromString(pathStr, ins.pathStr).Stat()
		if hash.String() != res.Hash.String() {
			err.Debug("Hash did not match", hash, res.Hash, pathStr)
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
