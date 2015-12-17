package collection

import (
	"github.com/adamcolton/err"
	"io"
	"math/rand"
	"os"
)

type Sync struct {
	a        *Instance
	b        *Instance
	actions  map[int][]Action
	maxDepth int
}

func (sync *Sync) Diff() {
	if roIdA, ok := sync.a.settings["readonly"]; ok && len(roIdA) == readOnlyIdLen {
		if roIdB, ok := sync.b.settings["readonly"]; !ok || len(roIdB) != readOnlyIdLen {
			// if a and b are both read only, no syncing
			sync.ReadOnlyDiff(sync.a, sync.b)
		}
		return
	}
	if roIdB, ok := sync.b.settings["readonly"]; ok && len(roIdB) == readOnlyIdLen {
		sync.ReadOnlyDiff(sync.b, sync.a)
		return
	}

	for id, aDir := range sync.a.directories {
		if bDir, ok := sync.b.directories[id]; ok {
			if i := aDir.PathNodes.DiffAt(bDir.PathNodes); i != -1 {
				err.Debug("ResDir", aDir.FullPath())
				sync.ResolveDirectoryDifference(id, i)
			}
		} else {
			err.Debug("CpyDir", aDir.FullPath())
			sync.MakeDirectory(aDir, sync.b)
			sync.b.dirty = true
		}
	}
	for id, bDir := range sync.b.directories {
		if _, ok := sync.a.directories[id]; !ok {
			err.Debug("CpyDir", bDir.FullPath())
			sync.MakeDirectory(bDir, sync.a)
			sync.a.dirty = true
		}
	}

	for id, aRes := range sync.a.resources {
		if bRes, ok := sync.b.resources[id]; ok {
			if i := aRes.PathNodes.DiffAt(bRes.PathNodes); i != -1 {
				err.Debug("Resolve", aRes.FullPath())
				sync.ResolveResourceDifference(id, i)
			}
		} else {
			err.Debug("Copy", aRes.FullPath())
			sync.CopyResource(aRes, sync.b)
			sync.b.dirty = true
		}
	}
	for id, bRes := range sync.b.resources {
		if _, ok := sync.a.resources[id]; !ok {
			err.Debug("Copy", bRes.FullPath())
			sync.CopyResource(bRes, sync.a)
			sync.a.dirty = true
		}
	}
}

func (sync *Sync) addAction(depth int, action Action) {
	actionList, ok := sync.actions[depth]
	if !ok {
		actionList = make([]Action, 0)
	}
	sync.actions[depth] = append(actionList, action)
	if depth > sync.maxDepth {
		sync.maxDepth = depth
	}
}

func (sync *Sync) Run() bool {
	l := len(sync.actions)
	if l == 0 {
		return false
	}

	for i := 0; i <= sync.maxDepth; i++ {
		sync.runList(i)
	}
	sync.runList(-1)

	return true
}
func (sync *Sync) runList(depth int) {
	actionList, ok := sync.actions[depth]
	if ok {
		for i := 0; i < len(actionList); i++ {
			actionList[i].Execute()
		}
	}
}

type Action interface {
	Execute()
}

type CpRes struct {
	res  *Resource
	ins  *Instance
	sync *Sync
}

func (sync *Sync) CopyResource(res *Resource, ins *Instance) {
	sync.addAction(res.Depth(), &CpRes{
		res:  res,
		ins:  ins,
		sync: sync,
	})
}

func (cpRes *CpRes) Execute() {
	err.Debug("Copying: ", cpRes.res.FullPath())
	if cpRes.res.PathNodes.Last().Parent() == nil {
		cpRes.copyResData()
	} else if cpRes.copyContents() {
		cpRes.copyResData()
	}
}

func (cpRes *CpRes) copyResData() {
	pns := &PathNodes{
		nodes: make([]*PathNode, len(cpRes.res.PathNodes.nodes)),
	}
	for i, srcN := range cpRes.res.PathNodes.nodes {
		pns.nodes[i] = &PathNode{
			Name:     srcN.Name,
			ParentID: srcN.ParentID,
			Instance: cpRes.ins,
		}
	}
	cpRes.ins.resources[cpRes.res.ID.String()] = &Resource{
		ID:        cpRes.res.ID,
		Hash:      cpRes.res.Hash,
		PathNodes: pns,
	}
}

func (cpRes *CpRes) copyContents() bool {
	srcStr := cpRes.res.FullPath()
	dstRelPath := cpRes.res.RelativePath()

	dstStrRoot := cpRes.ins.pathStr
	dstStrRoot += dstRelPath.relDir

	name, moved := confirmedAavailableName(dstStrRoot, dstRelPath.name)
	dstStr := dstStrRoot + name

	if srcFile, e := os.Open(srcStr); err.Log(e) {
		defer srcFile.Close()
		if dstFile, e := os.Create(dstStr); err.Log(e) {
			defer dstFile.Close()
			if _, e := io.Copy(dstFile, srcFile); err.Log(e) {
				dstFile.Sync()
				if moved {
					cpRes.sync.addAction(-1, &RetryRename{
						current: dstStr,
						target:  dstStrRoot + dstRelPath.name,
					})
				}
				return true
			}
		}
	}
	return false
}

type CpDir struct {
	dir  *Directory
	ins  *Instance
	sync *Sync
}

func (sync *Sync) MakeDirectory(dir *Directory, ins *Instance) {
	sync.addAction(dir.Depth(), &CpDir{
		dir:  dir,
		ins:  ins,
		sync: sync,
	})
}

func (cpDir *CpDir) Execute() {
	err.Debug("Copying: ", cpDir.dir.FullPath())
	dstStr := cpDir.ins.pathStr + cpDir.dir.RelativePath().String()
	os.Mkdir(dstStr, 0700)
	pns := &PathNodes{
		nodes: make([]*PathNode, len(cpDir.dir.PathNodes.nodes)),
	}
	for i, srcN := range cpDir.dir.PathNodes.nodes {
		pns.nodes[i] = &PathNode{
			Name:     srcN.Name,
			ParentID: srcN.ParentID,
			Instance: cpDir.ins,
		}
	}
	dir := &Directory{
		Resource: &Resource{
			ID:        cpDir.dir.ID,
			Hash:      cpDir.dir.Hash,
			PathNodes: pns,
		},
		directories: make(map[string]*Directory),
		resources:   make(map[string]*Resource),
	}
	cpDir.ins.directories[cpDir.dir.ID.String()] = dir
	pn := dir.PathNodes.Last()
	if parent := pn.Parent(); parent != nil {
		parent.directories[pn.Name] = dir
	} else {
		panic("Parent is nil")
	}
}

func (sync *Sync) ResolveDirectoryDifference(id string, divergeStart int) {
	a := sync.a.directories[id].Resource
	b := sync.b.directories[id].Resource
	depth, action := resolveDifference(a, b, divergeStart, sync)
	sync.addAction(depth, action)
}

// MvRes is actually move or delete, because we consider "deleted" to be
// a virtual location
//
// This keeps throwing me:
// make "To" like "From" -> only change "To"
type MvRes struct {
	cloneFrom *Resource
	cloneTo   *Resource
	start     int
	sync      *Sync
}

func (sync *Sync) ResolveResourceDifference(id string, divergeStart int) {
	a := sync.a.resources[id]
	b := sync.b.resources[id]
	depth, action := resolveDifference(a, b, divergeStart, sync)
	sync.addAction(depth, action)
}

func resolveDifference(a, b *Resource, divergeStart int, sync *Sync) (int, Action) {
	apn := a.PathNodes.Last()
	bpn := b.PathNodes.Last()

	a2b := &MvRes{
		cloneFrom: a,
		cloneTo:   b,
		start:     divergeStart,
		sync:      sync,
	}

	b2a := &MvRes{
		cloneFrom: b,
		cloneTo:   a,
		start:     divergeStart,
		sync:      sync,
	}

	mv := a2b

	// If divergeStart is equal to the length of one of the PathNode lists,
	// that has highest priority; they were syncd at one point and then
	// more has happened to the longer
	// The next priority is if one of them was deleted, we choose the other
	// If all else fails (neither is deleted and they have divergent histories)
	// we choose which ever one has the longest relative path.
	if len(a.PathNodes.nodes) == divergeStart {
		mv = b2a
	} else if len(b.PathNodes.nodes) == divergeStart || bpn.ParentID == nil {
		if bpn.ParentID == nil && bpn.Name != ".deleted" {
			err.Debug(bpn.Name)
			panic("Bad Node")
		}
	} else if apn.ParentID == nil {
		if apn.Name != ".deleted" {
			panic("Bad Node")
		}
		mv = b2a
	} else if len(bpn.RelativePath().String()) > len(apn.RelativePath().String()) {
		mv = b2a
	}

	mv.cloneTo.PathNodes.Last().Instance.dirty = true
	if pn := mv.cloneTo.PathNodes.Last(); pn.ParentID == nil && pn.Name == ".deleted" {
		return mv.cloneFrom.Depth(), Action(&CpRes{
			res:  mv.cloneFrom,
			ins:  pn.Instance,
			sync: sync,
		})
	}

	return mv.cloneFrom.Depth(), Action(mv)
}

func confirmedAavailableName(dirPath, name string) (string, bool) {
	availableName := name
	moved := false
	for {
		if _, err := os.Stat(dirPath + availableName); os.IsNotExist(err) {
			// if the file does not exist, exit the loop
			// return fullPath as the file name
			return availableName, moved
		}
		// if the file file exists
		// choose a ranodm modifier and prepend to the name
		mod := make([]rune, 7)
		for i := 1; i < 6; i++ {
			mod[i] = rune((rand.Float32() * 24) + 65)
		}
		mod[0] = '0'
		mod[6] = '_'
		availableName = string(mod) + name
		moved = true
	}
}

func (mvRes *MvRes) Execute() {
	cloneFromNode := mvRes.cloneFrom.PathNodes.Last()
	cloneToNode := mvRes.cloneTo.PathNodes.Last()

	if cloneFromNode.ParentID == nil {
		if cloneFromNode.Name != ".deleted" {
			panic("Bad Node")
		} else if e := os.RemoveAll(cloneToNode.FullPath()); err.Log(e) {
			copyNodes(mvRes.cloneFrom.PathNodes, mvRes.cloneTo.PathNodes, mvRes.start)
		}
	} else {
		//check that there isn't a file there, if there is, create a random prefix.
		cloneToStrRoot := cloneToNode.Instance.pathStr
		cloneToStrRoot += cloneFromNode.RelativePath().relDir
		name, moved := confirmedAavailableName(cloneToStrRoot, cloneFromNode.Name)
		cloneToStr := cloneToStrRoot + name
		if e := os.Rename(cloneToNode.FullPath(), cloneToStr); err.Log(e) {
			if moved {
				mvRes.sync.addAction(-1, &RetryRename{
					current: cloneToStr,
					target:  cloneToStrRoot + cloneFromNode.Name,
				})
			}
			copyNodes(mvRes.cloneFrom.PathNodes, mvRes.cloneTo.PathNodes, mvRes.start)
		}
	}
}

func copyNodes(fromPns, toPns *PathNodes, start int) {
	ins := toPns.Last().Instance //in theory, all in the instance values should be the same
	for i := start; i < len(fromPns.nodes); i++ {
		cloneFrom := fromPns.nodes[i]
		pn := &PathNode{
			Name:     cloneFrom.Name,
			ParentID: cloneFrom.ParentID,
			Instance: ins,
		}
		if i >= len(toPns.nodes) {
			toPns.Add(pn)
		} else {
			toPns.nodes[i] = pn
		}
	}
	if len(toPns.nodes) > len(fromPns.nodes) {
		toPns.nodes = toPns.nodes[:len(fromPns.nodes)]
	}
}

type RetryRename struct {
	current string
	target  string
}

func (retry *RetryRename) Execute() {
	if _, err := os.Stat(retry.target); os.IsNotExist(err) {
		os.Rename(retry.current, retry.target)
	}
}

func (sync *Sync) ReadOnlyDiff(readOnly, write *Instance) {
	for id, rDir := range readOnly.directories {
		if _, ok := sync.b.directories[id]; !ok {
			err.Debug("RO CpyDir", rDir.FullPath())
			sync.MakeDirectory(rDir, write)
			write.dirty = true
		}
	}

	for id, wRes := range readOnly.resources {
		if _, ok := sync.b.resources[id]; !ok {
			err.Debug("RO Copy", wRes.FullPath())
			sync.CopyResource(wRes, write)
			write.dirty = true
		}
	}
}
