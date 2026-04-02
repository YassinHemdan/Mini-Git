package internals

import (
	"path/filepath"
	"strings"
)

type Blob struct {
	oid         []byte
	data        []byte
	pathname    string
	permissions string
}

func (b *Blob) New(data []byte, pathname string, isExecutable bool) error {
	b.data = data
	b.pathname = pathname
	b.permissions = "100755"
	if !isExecutable {
		b.permissions = "100644"
	}
	return nil
}
func (b *Blob) ToString() string {
	return string(b.data)
}

func (b *Blob) GetOid() []byte {
	return b.oid
}

func (b *Blob) SetOid(oid []byte) {
	b.oid = oid
}

func (b *Blob) Type() string {
	return "blob"
}

func (b *Blob) GetName() string {
	return filepath.Base(b.pathname)
}

func (b *Blob) GetPathname() string {
	return b.pathname
}

func (b *Blob) GetMode() string {
	return b.permissions
}

func (e *Blob) ParentDirectories() []string {
	prefixs := strings.Split(filepath.ToSlash(e.GetPathname()), "/")
	parents := []string{}

	for i := 1; i < len(prefixs); i++ {
		parents = append(parents, strings.Join(prefixs[:i], "/"))
	}

	return parents
}

