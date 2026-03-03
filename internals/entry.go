package internals

import (
	"fmt"
	"io/fs"
)

type Entry struct {
	oid  []byte
	name string
	mode fs.FileMode
}

func (e *Entry) New(oid []byte, name string, mode fs.FileMode) error {
	if len(oid) == 0 || len(name) == 0 {
		return fmt.Errorf("name or oid cannot be empty")
	}
	e.name = name
	e.oid = oid
	e.mode = mode
	return nil
}

func (e *Entry) GetOid() []byte {
	return e.oid
}
func (e *Entry) SetOid(oid []byte) {
	e.oid = oid
}
func (e *Entry) GetName() string {
	return e.name
}
func (e *Entry) GetMode() fs.FileMode {
	return e.mode
}
