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
func (b *Tree) Type() string {
	return "tree"
}
func (t *Tree) GetOid() []byte {
	return t.oid
}

func (t *Tree) SetOid(oid []byte) {
	t.oid = oid
}

func (t *Tree) ToString() string {
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

	var entries []byte
	for _, entry := range data_copy {
		entries = append(entries, fmt.Sprintf("10%04o %s%x", entry.GetMode(), entry.GetName(), 0x00)...)
		entries = append(entries, entry.GetOid()...)
	}
	return string(entries)
}
