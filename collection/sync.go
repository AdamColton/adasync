package collection

import (
	"github.com/adamcolton/err"
	"io"
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

func (cpDir *CpDir) Execute() {
	dstStr := cpDir.ins.pathStr + cpDir.dir.RelativePath().String()
	if e := os.Mkdir(dstStr, 0700); err.Log(e) {
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
}

type dirToP struct {
	dir *Directory
}

func (dtp *dirToP) Execute() {
	//After everything is copied, we sync up directories.
	dtp.dir.PathNodes.Last().Parent().directories[dtp.dir.ID.String()] = dtp.dir
}

func (sync *Sync) ResolveDirectoryDifference(id string, startResolve int) {
	return
	a := sync.a.resources[id]
	//b := sync.b.resources[id]
	masterPn := &PathNodes{
		nodes: make([]*PathNode, startResolve),
	}
	for i := 0; i < startResolve; i++ {
		n := a.PathNodes.nodes[i]
		masterPn.nodes[i] = &PathNode{
			Name:     n.Name,
			ParentID: n.ParentID,
		}
	}
}

func (sync *Sync) ResolveResourceDifference(id string, i int) {

}
