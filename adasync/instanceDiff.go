package adasync

import (
	"github.com/adamcolton/err"
	"os"
	"path/filepath"
	"sort"
)

func (ins *Instance) SelfUpdate() {
	err.Debug("Self Update: ", ins.pathStr)
	diff := ins.SelfDiff()
	// directories need to be resolved first, otherwise if a directory was
	// renamed, every file will think it was moved.
	diff.resolveDirectories()
	diff.resolveFiles()
	diff.resolveDeleted()
}

// SelfDiff
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
	return diff
}

//both are used as sets, not maps
type deltaSelf struct {
	added         []string
	removed       map[string]*Resource // path -> resource
	removedByHash map[string]*Resource // hash -> resource
	ins           *Instance
	deleted       map[string]*Resource // id -> resource
}

// addDirs is used to walk the directory
func (d *deltaSelf) addDirs(pathStr string, fi os.FileInfo, _ error) error {
	if !fi.IsDir() {
		return nil
	}

	// sanitize pathStr
	pathStr = PathFromString(endingSlash(pathStr), d.ins.pathStr).String()

	d.checkFile(pathStr)
	return nil
}

// add adds a path to a deltaSelf. If the path is in "removed"
// then it's a known resource and it's removed from removed
// if not, then it's a new resources and is added to added.
func (d *deltaSelf) addFiles(pathStr string, fi os.FileInfo, _ error) error {
	if fi.IsDir() {
		return nil
	}
	path := PathFromString(pathStr, d.ins.pathStr)
	// skip any files ending in .collection
	if endsWith(path.name, ".collection") {
		return nil
	}
	pathStr = path.String()
	d.checkFile(pathStr)
	return nil
}

func (d *deltaSelf) checkFile(pathStr string) {
	if res, ok := d.removed[pathStr]; ok && d.ins.pathEqualsResource(res, pathStr) {
		delete(d.removed, pathStr)
		delete(d.removedByHash, res.Hash.String())
	} else {
		d.added = append(d.added, pathStr)
	}
}

func (d *deltaSelf) resolveDeleted() {
	for _, res := range d.removed {
		d.ins.dirty = true
		res.PathNodes.Add(d.ins.PathNodeFromHash(nil, ".deleted"))
	}
}

func (d *deltaSelf) resolveDirectories() {
	// Put everything in removed and remove from removed
	// as we find each. What's left is what was actually
	// removed
	for _, dir := range d.ins.directories {
		if pn := dir.PathNodes.Last(); pn.ParentID != nil || pn.Name != ".deleted" {
			d.removed[pn.FullPath()] = dir.Resource
			d.removedByHash[dir.ID.String()] = dir.Resource
		} else if pn.Name == ".deleted" {
			d.deleted[dir.ID.String()] = dir.Resource
		}
	}

	d.added = make([]string, 0)
	filepath.Walk(d.ins.pathStr, d.addDirs)

	// We sort so that files will be added in an order such that a child can
	// always add itself to it's parent. But there may be cases involving moving
	// where that won't work
	sort.Sort(ByLength(d.added))

	for h, res := range d.removedByHash {
		err.Debug(h, res.RelativePath())
	}
	for _, newPathStr := range d.added {
		d.ins.dirty = true
		newPath := PathFromString(newPathStr, d.ins.pathStr) //*Path
		pathNode := d.ins.PathToNode(newPath)                //*PathNode
		hash, _, _ := newPath.Stat()
		err.Debug(hash, newPathStr)
		if res, ok := d.removedByHash[hash.String()]; ok {

			// resource was moved
			err.Debug("Moved: ", res.FullPath())
			err.Debug("To: ", newPathStr)
			delete(d.removed, res.FullPath())
			delete(d.removedByHash, hash.String())

			parent := res.PathNodes.Last().Parent()
			delete(parent.directories, res.RelativePath().name)
			res.PathNodes.Add(pathNode)
			parent.directories[res.RelativePath().name] = d.ins.directories[res.ID.String()]
		} else {
			// resource is new
			err.Debug("Added: ", newPathStr)
			dir := d.ins.AddDirectoryWithPath(hash, pathNode)
			dir.PathNodes.Last().Parent().directories[dir.RelativePath().name] = dir
		}

	}
}

func (d *deltaSelf) resolveFiles() {
	// Put everything in removed and remove from removed
	// as we find each. What's left is what was actually
	// removed
	for _, res := range d.ins.resources {
		if pn := res.PathNodes.Last(); pn.ParentID != nil || pn.Name != ".deleted" {
			d.removed[pn.FullPath()] = res
			d.removedByHash[res.Hash.String()] = res
		} else if pn.Name == ".deleted" {
			d.deleted[res.ID.String()] = res
		}
	}

	d.added = make([]string, 0)
	filepath.Walk(d.ins.pathStr, d.addFiles)

	for _, newPathStr := range d.added {
		d.ins.dirty = true
		newPath := PathFromString(newPathStr, d.ins.pathStr) //*Path
		pathNode := d.ins.PathToNode(newPath)                //*PathNode
		hash, _, size := newPath.Stat()
		if res, ok := d.removedByHash[hash.String()]; ok {
			// resource was moved
			err.Debug("Moved: ", res.FullPath())
			err.Debug("To: ", newPathStr)
			delete(d.removed, res.FullPath())
			delete(d.removedByHash, hash.String())
			res.PathNodes.Add(pathNode)
		} else {
			// resource is new
			err.Debug("Added: ", newPathStr)
			r := d.ins.AddResourceWithPath(hash, size, pathNode)
			err.Debug(r.Size, size)
		}
	}
}

// BadInstanceScan this is a debugging tool
// despite my best efforts, unit testing has not caught all the errors, this
// can help find additional errors under real conditions
func (ins *Instance) BadInstanceScan() {

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

}
