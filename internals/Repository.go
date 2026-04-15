package internals

import (
	index "JIT/internals/index"
	"path/filepath"
)

type Repository struct {
	path      string
	refs      *Refs
	index     *index.Index
	database  *Database
	workspace *Workspace
}

func NewRepository(pathname string) (*Repository, error) {
	index, err := index.NewIndex(filepath.Join(pathname, "index"))
	if err != nil {
		return nil, err
	}

	refs, err := NewRefs(pathname)
	if err != nil {
		return nil, err
	}
	db, err := NewDatabase(filepath.Join(pathname, "objects"))
	if err != nil {
		return nil, err
	}

	workspace, err := NewWorkspace(filepath.Dir(pathname))
	if err != nil {
		return nil, err
	}
	return &Repository{
		path:      pathname,
		index:     index,
		refs:      refs,
		database:  db,
		workspace: workspace,
	}, nil
}

func (r *Repository) Index() *index.Index {
	return r.index
}

func (r *Repository) Refs() *Refs {
	return r.refs
}

func (r *Repository) Workspace() *Workspace {
	return r.workspace
}

func (r *Repository) Database() *Database {
	return r.database
}
