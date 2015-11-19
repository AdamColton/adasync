package collection

import (
	"github.com/adamcolton/err"
	"io"
	"math/rand"
	"os"
)

type Sync struct {
	a       *Instance
	b       *Instance
	actions []Action
}

func (sync *Sync) Diff() {
	for id, aDir := range sync.a.directories {
		if bDir, ok := sync.b.directories[id]; ok {
			if i := aDir.PathNodes.DiffAt(bDir.PathNodes); i != -1 {
				sync.ResolveDirectoryDifference(id, i)
			}
		} else {
			sync.MakeDirectory(aDir, sync.b)
		}
	}
	for id, bDir := range sync.b.directories {
		if _, ok := sync.a.directories[id]; !ok {
			sync.MakeDirectory(bDir, sync.a)
		}
	}

	for id, aRes := range sync.a.resources {
		if bRes, ok := sync.b.resources[id]; ok {
			if i := aRes.PathNodes.DiffAt(bRes.PathNodes); i != -1 {
				sync.ResolveResourceDifference(id, i)
			}
		} else {
			sync.CopyResource(aRes, sync.b)
		}
	}
	for id, bRes := range sync.b.resources {
		if _, ok := sync.a.resources[id]; !ok {
			sync.CopyResource(bRes, sync.a)
		}
	}
}

func (sync *Sync) Run() {
	for i := 0; i < len(sync.actions); i++ {
		sync.actions[i].Execute()
	}
}

type Action interface {
	Execute()
}

type CpRes struct {
	res *Resource
	ins *Instance
}

func (sync *Sync) CopyResource(res *Resource, ins *Instance) {
	sync.actions = append(sync.actions, &CpRes{
		res: res,
		ins: ins,
	})
}

func (cpRes *CpRes) Execute() {
	if cpRes.res.PathNodes.Last().Parent() == nil {
		//TODO: Delete
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
	dstStr := cpRes.ins.pathStr + cpRes.res.RelativePath().String()

	if srcFile, e := os.Open(srcStr); err.Log(e) {
		defer srcFile.Close()
		if dstFile, e := os.Create(dstStr); err.Log(e) {
			defer dstFile.Close()
			if _, e := io.Copy(dstFile, srcFile); err.Log(e) {
				dstFile.Sync()
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
	sync.actions = append(sync.actions, &CpDir{
		dir:  dir,
		ins:  ins,
		sync: sync,
	})
}

// RESUME: only Mkdir if dir does not exist.
func (cpDir *CpDir) Execute() {
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
	cpDir.sync.actions = append(cpDir.sync.actions, &dirToP{dir: dir})
}

type dirToP struct {
	dir *Directory
}

func (dtp *dirToP) Execute() {
	//After everything is copied, we sync up directories.
	//I think we can do this earlier
	dtp.dir.PathNodes.Last().Parent().directories[dtp.dir.ID.String()] = dtp.dir
}

func (sync *Sync) ResolveDirectoryDifference(id string, divergeStart int) {
	a := sync.a.directories[id].Resource
	b := sync.b.directories[id].Resource
	sync.actions = append(sync.actions, resolveDifference(a, b, divergeStart))
}

// MvRes is actually move or delete, because we consider "deleted to be
// a virtual location
type MvRes struct {
	cloneFrom *Resource
	cloneTo   *Resource
	start     int
}

func (sync *Sync) ResolveResourceDifference(id string, divergeStart int) {
	a := sync.a.resources[id]
	b := sync.b.resources[id]
	sync.actions = append(sync.actions, resolveDifference(a, b, divergeStart))
}

func resolveDifference(a, b *Resource, divergeStart int) Action {
	apn := a.PathNodes.Last()
	bpn := b.PathNodes.Last()

	a2b := &MvRes{
		cloneFrom: a,
		cloneTo:   b,
		start:     divergeStart,
	}

	b2a := &MvRes{
		cloneFrom: b,
		cloneTo:   a,
		start:     divergeStart,
	}

	action := Action(a2b)

	// If divergeStart is equal to the length of one of the PathNode lists,
	// that has highest priority; they were syncd at one point and then
	// more has happened to the longer
	// The next priority is if one of them was deleted, we choose the other
	// If all else fails (neither is deleted and they have divergent histories)
	// we choose which ever one has the longest relative path.
	if len(a.PathNodes.nodes) == divergeStart {
		action = b2a
	} else if len(b.PathNodes.nodes) == divergeStart || bpn.ParentID == nil {
		if bpn.ParentID == nil && bpn.Name != ".deleted" {
			panic("Bad Node")
		}
	} else if apn.ParentID == nil {
		if apn.Name != ".deleted" {
			panic("Bad Node")
		}
		action = b2a
	} else if len(bpn.RelativePath().String()) > len(apn.RelativePath().String()) {
		action = b2a
	}
	return action
}

func (mvRes *MvRes) Execute() {
	cloneFromStr := mvRes.cloneFrom.FullPath()
	cloneToNode := mvRes.cloneTo.PathNodes.Last()

	if cloneToNode.ParentID == nil {
		if cloneToNode.Name != ".deleted" {
			panic("Bad Node")
		} else if e := os.Remove(cloneFromStr); err.Log(e) {
			copyNodes(mvRes.cloneFrom.PathNodes, mvRes.cloneTo.PathNodes, mvRes.start)
		}
	} else {
		//check that there isn't a file there, if there is, create a random prefix.
		cloneToStrRoot := cloneToNode.Parent().FullPath()
		name := cloneToNode.Name
		cloneToStr := cloneToStrRoot + name
		for {
			if _, err := os.Stat(cloneToStr); os.IsNotExist(err) {
				// file exists
				mod := make([]rune, 7)
				for i := 1; i < 6; i++ {
					mod[i] = rune((rand.Float32() * 24) + 65)
				}
				mod[0] = '0'
				mod[6] = '_'
				cloneToStr = cloneToStrRoot + string(mod) + name
			} else {
				break
			}
		}
		if e := os.Rename(cloneFromStr, cloneToStr); err.Log(e) {
			copyNodes(mvRes.cloneFrom.PathNodes, mvRes.cloneTo.PathNodes, mvRes.start)
		}
	}
}

func copyNodes(fromPns, toPns *PathNodes, start int) {
	ins := fromPns.Last().Instance //in theory, all in the instance values should be the same
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
}
