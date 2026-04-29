package internals

import (
	"JIT/internals/utils"
	scanner "JIT/utils"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
)

type Tree struct {
	oid      []byte
	entries  map[string]Entry
	pathname string
	keys     []string
}

func NewTree(entries map[string]Entry) *Tree {
	var treeEntries map[string]Entry
	if entries == nil {
		treeEntries = make(map[string]Entry)
	} else {
		treeEntries = entries
	}

	return &Tree{
		entries: treeEntries,
	}
}
func ParseTree(scanner *scanner.SmartScanner) Object {
	oidSplitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		start := 0
		if start < len(data) && data[start] == '\n' {
			return start + 1, data[start : start+1], nil
		}
		return start + 20, data[start : start+20], nil
	}

	treeEntries := make(map[string]Entry)

	scanner.SplitByDelim(' ', true)
	for scanner.Scan() {
		readedMode := scanner.Text()
		mode, _ := strconv.ParseUint(readedMode, 8, 32)

		scanner.SplitByDelim('\x00', true)
		scanner.Scan()
		entryName := scanner.Text()

		scanner.SetSplit(oidSplitFunc)
		scanner.Scan()
		entryOid := fmt.Sprintf("%x", scanner.Text())
		decodedOid, _ := hex.DecodeString(entryOid)

		treeEntry := NewTreeEntry(entryName, decodedOid, uint32(mode))

		treeEntries[entryName] = treeEntry
		scanner.SplitByDelim(' ', true)
	}
	return NewTree(treeEntries)

}

func (t *Tree) AddEntry(ParentDirectories []string, entry Entry) {
	if len(ParentDirectories) == 0 {
		t.entries[entry.GetName()] = entry
		t.keys = append(t.keys, entry.GetName())
	} else {
		childTree := NewTree(nil)
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

func BuildTree(entries []BuildEntry) *Tree {
	root := NewTree(nil)
	for _, entry := range entries {
		root.AddEntry(entry.ParentDirectories(), entry)
	}
	return root
}

/*
Notice here that we are only saving the trees only
We are not checkinf for the blobs as they are already saved during the staging phase
We are only care about the OIDs of a tree's entries to serilaize it and save it in the DB
*/
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
}

func (t *Tree) SetTreePathname(pathname string) {
	t.pathname = pathname
}
func (t *Tree) GetPathname() string {
	return t.pathname
}
func (t *Tree) GetEntries() map[string]Entry {
	return t.entries
}
