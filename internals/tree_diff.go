package internals

import (
	internals "JIT/internals/database"
	"fmt"
	"path/filepath"
)

type treeDiff struct {
	database *Database
	changes  map[string][]internals.Entry
}

func newTreeDiff(database *Database) *treeDiff {
	return &treeDiff{database: database, changes: make(map[string][]internals.Entry)}
}

func (t *treeDiff) CompareOIDs(a_oid, b_oid []byte, path string) error {
	if string(a_oid) == string(b_oid) {
		return nil
	}
	tree_a, err := t.oidToTree(a_oid)
	if err != nil {
		return err
	}

	tree_b, err := t.oidToTree(b_oid)
	if err != nil {
		return err
	}

	treeEntires := func(treeObj *internals.Tree) map[string]internals.Entry {
		if treeObj == nil {
			return make(map[string]internals.Entry, 0)
		}
		return treeObj.GetEntries()
	}

	a_entries, b_entries := treeEntires(tree_a), treeEntires(tree_b)
	if err := t.detectDeletions(a_entries, b_entries, path); err != nil {
		return err
	}
	if err := t.detectAdditions(a_entries, b_entries, path); err != nil {
		return err
	}

	return nil
}

func (t *treeDiff) detectDeletions(a_entries, b_entries map[string]internals.Entry, path string) error {
	// in fact, this method handles deleted and changed files
	for name, entry := range a_entries {
		other, found := b_entries[name]
		fullpath := filepath.Join(path, name)
		if found && entry.GetMode() == other.GetMode() && string(entry.GetOid()) == string(other.GetOid()) {
			continue
		} else {
			treeOid := func(entry internals.Entry) []byte {
				if entry == nil || entry.Type() != "tree" {
					return nil
				}
				return entry.GetOid()
			}
			getBlob := func(entry internals.Entry) internals.Entry {
				if entry == nil || entry.Type() == "tree" {
					return nil
				}
				return entry
			}

			if err := t.CompareOIDs(treeOid(entry), treeOid(other), fullpath); err != nil {
				return err
			}

			blobs := make([]internals.Entry, 0)
			blobs = append(blobs, getBlob(entry))
			blobs = append(blobs, getBlob(other))

			if blobs[0] != nil || blobs[1] != nil {
				t.changes[fullpath] = blobs
			}

		}
	}

	return nil
}
func (t *treeDiff) detectAdditions(a_entries, b_entries map[string]internals.Entry, path string) error {
	for name, entry := range b_entries {
		fullpath := filepath.Join(path, name)
		other, _ := a_entries[name]

		if other != nil { // if we found it, it would be already handled before by the detectDeletions
			continue
		}

		if entry.Type() == "tree" {
			if err := t.CompareOIDs(nil, entry.GetOid(), fullpath); err != nil {
				return err
			}
		} else {
			t.changes[fullpath] = []internals.Entry{nil, entry}
		}
	}
	return nil
}

func (t *treeDiff) oidToTree(oid []byte) (*internals.Tree, error) {
	if oid == nil {
		return nil, nil
	}
	object, err := t.database.Load(oid)
	if err != nil {
		return nil, err
	}

	if object.Type() == "commit" {
		commit := object.(*internals.Commit)
		return t.oidToTree(commit.GetTreeOid())
	} else if object.Type() == "tree" {
		return object.(*internals.Tree), nil
	}
	return nil, fmt.Errorf("Could not load tree object")
}

func (t *treeDiff) GetChanges() map[string][]internals.Entry {
	return t.changes
}
