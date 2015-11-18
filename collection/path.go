package collection

import (
	"crypto/md5"
	"github.com/adamcolton/err"
	"os"
	"path/filepath"
	"strings"
)

// Path
// instance should never end in /
// relative should always begin and end with /
// if this is a directory, it should end with /
type Path struct {
	root   string
	relDir string
	name   string
}

func (p *Path) String() string {
	return p.root + p.relDir + p.name
}

func PathFromString(fullpathStr, root string) *Path {
	fullpathStr = filepath.ToSlash(fullpathStr)
	root = filepath.ToSlash(root)
	relPathStr := strings.Replace(fullpathStr, root, "", -1)
	dir, name := split(relPathStr)
	return &Path{
		root:   root,
		relDir: dir,
		name:   name,
	}
}

func split(path string) (string, string) {
	addSlash := ""
	if l := len(path); l > 0 && path[l-1] == '/' {
		path = path[:l-1]
		addSlash = "/"
	}
	dir, name := filepath.Split(path)
	return dir, name + addSlash
}

var blocksize = int64(md5.BlockSize)

//MD5 efficiently finds the MD5 hash of the file at path
func (p *Path) Stat() (*Hash, bool) {
	file, e := os.Open(p.String())
	err.Panic(e)
	defer file.Close()
	stat, e := file.Stat()
	err.Panic(e)
	if stat.IsDir() {
		ret := Hash(md5.Sum([]byte(p.relDir + p.name)))
		return &ret, true
	}
	hash := md5.New()
	blocks := stat.Size() / blocksize
	if stat.Size()%blocksize != 0 {
		blocks++
	}

	buf := make([]byte, hash.BlockSize())
	for i := int64(0); i < blocks; i++ {
		l, e := file.Read(buf)
		err.Log(e)
		_, e = hash.Write(buf[:l])
		err.Warn(e)
	}
	sliceHash := hash.Sum(nil)
	ret := Hash{}
	for i, b := range sliceHash {
		ret[i] = b
	}
	return &ret, false
}

type PathNode struct {
	Name     string
	ParentID *Hash
	Instance *Instance
}

type PathNodes struct {
	nodes []*PathNode
}

func NewPathNodes(l int, pns ...*PathNode) *PathNodes {
	if len(pns) > 0 {
		return &PathNodes{nodes: pns}
	}
	return &PathNodes{nodes: make([]*PathNode, l)}
}

func (pns *PathNodes) Last() *PathNode {
	l := len(pns.nodes)
	if l == 0 {
		return nil
	}
	return pns.nodes[l-1]
}

func (pns *PathNodes) Add(pn ...*PathNode) {
	pns.nodes = append(pns.nodes, pn...)
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

func (pn *PathNode) IsDeleted() bool {
	return pn.ParentID == nil && pn.Name == ".deleted"
}

func (pns *PathNodes) marshal() []*SerialPathNode {
	sPns := make([]*SerialPathNode, len(pns.nodes))
	for i, pn := range pns.nodes {
		sPns[i] = pn.Serialize()
	}
	return sPns
}

func (ins *Instance) PathNodeFromHash(parentID *Hash, name string) *PathNode {
	return &PathNode{
		Name:     name,
		ParentID: parentID,
		Instance: ins,
	}
}

func (pn *PathNode) Parent() *Directory {
	if pn.ParentID == nil || pn.Instance == nil {
		return nil
	}
	parent, _ := pn.Instance.directories[pn.ParentID.String()]
	return parent
}

func (a *PathNodes) DiffAt(b *PathNodes) int {
	la := len(a.nodes)
	lb := len(b.nodes)
	ret := -1
	l := la
	if la != lb {
		if la > lb {
			l = lb
		}
		ret = l
	}
	for i := 0; i < l; i++ {
		na := a.nodes[i]
		nb := a.nodes[i]
		if na.Name != nb.Name || !na.ParentID.Equal(nb.ParentID) {
			return i
		}
	}
	return ret
}

func (pn *PathNode) RelativePath() *Path {
	var path *Path

	if parent := pn.Parent(); parent != nil {
		if parentPath := parent.PathNodes.Last(); parent == parentPath.Parent() {
			panic("Recursive loop" + pn.Name)
		}
		path = parent.RelativePath()
		path.relDir += path.name
	} else {
		path = &Path{}
	}
	path.name = pn.Name
	return path
}

func (pn *PathNode) FullPath() string {
	return pn.Instance.pathStr + pn.RelativePath().String()
}

func (pn *PathNode) Serialize() *SerialPathNode {
	parentID := &Hash{}
	parent := pn.Parent()
	if parent != nil {
		parentID = parent.ID
	}
	return &SerialPathNode{
		Name:     []byte(pn.Name),
		ParentID: parentID[:],
	}
}

func (spn *SerialPathNode) unmarshal(ins *Instance) *PathNode {
	return &PathNode{
		Name: string(spn.Name),
	}
}

type ByLength []string

func (a ByLength) Len() int           { return len(a) }
func (a ByLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByLength) Less(i, j int) bool { return len(a[i]) < len(a[j]) }