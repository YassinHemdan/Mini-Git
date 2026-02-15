package internals

import (
	"fmt"
	"slices"
)

type Tree struct {
	oid  []byte
	data []Entry
}

func (t *Tree) New(data []Entry) error {
	t.data = data

	return nil
}

func (t *Tree) GetOid() []byte {
	return t.oid
}

func (t *Tree) SetOid(oid []byte) {
	t.oid = oid
}

func (t *Tree) ToString() string {
	// we need to serialize our tree in a specific way to use it when storing it

	data_copy := make([]Entry, len(t.data))
	copy(data_copy, t.data)

	slices.SortFunc(data_copy, func(e1, e2 Entry) int {
		if e1.GetName() < e2.GetName() {
			return -1
		} else if e1.GetName() > e2.GetName() {
			return 1
		}
		return 0
	})

	// data => <type> <size>\x00<mode> <name>\x00<oid><mode> <name>\x00<oid><mode> <name>\x00<oid>....

	var entries []byte
	for _, entry := range data_copy {
		// Example: 100644 blob.go\x00f2a37562jhfh23782732ghj8
		// Example: 100644 cmd\x00f2a37562jhfh23782732ghj8
		entries = append(entries, fmt.Sprintf("10%04o %s%x", JitDefaultPermission, entry.GetName(), 0x00)...)
		entries = append(entries, entry.GetOid()...)
	}

	// fmt.Println(len(string(entries) + "*internals.Tree 312"))
	return string(entries)
}
