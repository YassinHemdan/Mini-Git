package internals

import (
	"JIT/utils"
)

type Blob struct {
	oid  []byte
	data []byte
}

func NewBlob(data []byte) *Blob {
	return &Blob{
		data: data,
	}
}
func ParseBlob(scanner *utils.SmartScanner) Object {
	scanner.ScanRest()
	data := scanner.Text()

	return NewBlob([]byte(data))
}

func (b *Blob) ToString() string {
	return string(b.data)
}

func (b *Blob) GetData() []byte {
	return b.data
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
