package internals

import (
	internals "JIT/internals/database"
	utils "JIT/internals/utils"
	ds "JIT/utils/datastructures"
)

const (
	MIGRATION_CREATE = "create"
	MIGRATION_DELETE = "delete"
	MIGRATION_UPDATE = "update"
)

type Migration struct {
	repo    *Repository
	changes map[string][]*ds.Pair[string, internals.Entry] // <action, <pathname, entry>>
	diff    map[string][]internals.Entry                   // <pathname, [entries]>
	mkdirs  map[string]struct{}
	rmdirs  map[string]struct{}
}

func newMigration(repo *Repository, diff map[string][]internals.Entry) *Migration {
	changes := make(map[string][]*ds.Pair[string, internals.Entry])
	// changes[MIGRATION_CREATE] = make([]*ds.Pair[string, internals.Entry], 0)
	// changes[MIGRATION_DELETE] = make([]*ds.Pair[string, internals.Entry], 0)
	// changes[MIGRATION_UPDATE] = make([]*ds.Pair[string, internals.Entry], 0)

	return &Migration{
		repo:    repo,
		diff:    diff,
		changes: changes,
		mkdirs:  make(map[string]struct{}),
		rmdirs:  make(map[string]struct{}),
	}
}

func (m *Migration) ApplyChanges() error {
	m.planChanges()
	if err := m.updateWorkspace(); err != nil {
		return err
	}
	return nil
}
func (m *Migration) planChanges() {
	// we need to figure our what to delete and what to create
	for pathname, entries := range m.diff {
		m.recordChange(pathname, entries[0], entries[1])
	}
}

func (m *Migration) recordChange(pathname string, oldEntry, newEntry internals.Entry) {
	var action string
	parentDirectores := utils.ParentDirectories(pathname)
	if oldEntry == nil { // then, will be created

		action = MIGRATION_CREATE
		// we need to get every parent that should be created before we create our new file
		for _, parentDirectory := range parentDirectores {
			m.mkdirs[parentDirectory] = struct{}{}
		}
	} else if newEntry == nil { // will be deleted
		action = MIGRATION_DELETE

		// we need to get every parent that might be deleted if it got empty
		for _, parentDirectory := range parentDirectores {
			m.rmdirs[parentDirectory] = struct{}{}
		}
	} else {
		action = MIGRATION_UPDATE
		for _, parentDirectory := range parentDirectores {
			m.mkdirs[parentDirectory] = struct{}{}
		}
	}

	// we always take the new entry
	// if the current action is a DELETE, we won't need an entry to get any extra information
	// so we can store a nil in this case\

	m.changes[action] = append(m.changes[action], &ds.Pair[string, internals.Entry]{
		First:  pathname,
		Second: newEntry,
	})
}

// we will delegate updating the workspace to the workspace class itself to deal with files
// the workspace is the one responsible for listing and making modifications to the file system
func (m *Migration) updateWorkspace() error {
	return m.repo.Workspace().applyMigration(m)
}

func (m *Migration) blobData(oid []byte) ([]byte, error) {
	blob, err := m.repo.Database().Load(oid)
	if err != nil {
		return nil, err
	}
	return blob.(*internals.Blob).GetData(), nil
}
