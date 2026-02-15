package internals

import "fmt"

type Entry struct {
	oid  []byte
	name string
}

func (e *Entry) New(oid []byte, name string) error {
	if len(oid) == 0 || len(name) == 0 {
		return fmt.Errorf("name or oid cannot be empty")
	}
	e.name = name
	e.oid = oid

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
