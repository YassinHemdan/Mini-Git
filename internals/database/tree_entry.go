package internals

import "fmt"


/*
	- We only need the name, oid, and the mode for any entry in our tree
	  So we will introduce the TreeEntry to carry this information for us
	
	- This is different from the IndexEntry and BuildEntry
*/
type TreeEntry struct {
	name string
	mode uint32
	oid  []byte
}

func NewTreeEntry(name string, oid []byte, mode uint32) *TreeEntry {
	return &TreeEntry{
		name: name,
		oid:  oid,
		mode: mode,
	}
}
func (te *TreeEntry) GetName() string {
	return te.name
}
func (te *TreeEntry) Type() string {
	if te.mode == 040000 {
		return "tree"
	}
	return "blob"
}
func (te *TreeEntry) GetMode() string {
	return fmt.Sprintf("%o", te.mode)
}

func (te *TreeEntry) GetOid() []byte {
	return te.oid
}
