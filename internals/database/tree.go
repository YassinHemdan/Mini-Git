package internals

import (
	"JIT/internals/utils"
	"fmt"
	"path/filepath"
)

type Tree struct {
	oid      []byte
	entries  map[string]Entry
	pathname string
	keys     []string
}

func (t *Tree) New() {
	t.entries = make(map[string]Entry, 0)
}

func (t *Tree) AddEntry(ParentDirectories []string, entry Entry) {
	if len(ParentDirectories) == 0 {
		t.entries[entry.GetName()] = entry
		t.keys = append(t.keys, entry.GetName())
	} else {
		childTree := &Tree{}
		childTree.New()
		val, ok := t.entries[filepath.Base(ParentDirectories[0])]
		if ok {
			childTree = val.(*Tree) // already saved before, use it
		}

		childTree.SetTreePathname(ParentDirectories[0])
		childTree.AddEntry(ParentDirectories[1:], entry)

		t.entries[childTree.GetName()] = childTree
		if !ok {
			t.keys = append(t.keys, childTree.GetName())
		}
	}
}

func BuildTree(entries []Entry) *Tree {
	root := Tree{}
	root.New()
	for _, entry := range entries {
		root.AddEntry(entry.ParentDirectories(), entry)
	}
	return &root
}

func (t *Tree) Traverse(fn func(*Tree)) {
	for _, k := range t.keys {
		entry := t.entries[k]
		if entry.Type() == "tree" {
			childTree := entry.(*Tree)
			childTree.Traverse(fn)
		}
	}
	fn(t)
}

func (t *Tree) Type() string {
	return "tree"
}
func (t *Tree) GetOid() []byte {
	return t.oid
}

func (t *Tree) SetOid(oid []byte) {
	t.oid = oid
}

func (t *Tree) ToString() string {
	var data []byte
	for _, k := range t.keys {
		curEntry := t.entries[k]
		data = append(data, fmt.Sprintf("%s %s", curEntry.GetMode(), curEntry.GetName())...)
		data = append(data, 0x00)
		data = append(data, curEntry.GetOid()...)
	}
	return string(data)
}

func (t *Tree) GetName() string {
	return filepath.Base(t.pathname)
}

func (t *Tree) GetMode() string {
	return "40000"
}

func (t *Tree) ParentDirectories() []string {
	return utils.ParentDirectories(t.GetPathname())
	// prefixs := strings.Split(filepath.ToSlash(t.GetPathname()), "/")
	// parents := []string{}

	// for i := 1; i < len(prefixs); i++ {
	// 	parents = append(parents, strings.Join(prefixs[:i], "/"))
	// }

	// return parents
}

func (t *Tree) SetTreePathname(pathname string) {
	t.pathname = pathname
}
func (t *Tree) GetPathname() string {
	return t.pathname
}
