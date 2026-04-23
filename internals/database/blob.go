package internals

type Blob struct {
	oid  []byte
	data []byte
}

func NewBlob(data []byte) *Blob {
	return &Blob{
		data: data,
	}
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
