package internals

type Blob struct {
	oid  []byte // object id --> the content of the blob will be hashed to generate the oid
	data []byte // the content of the blob

}

func (b *Blob) New(data []byte) error {
	b.data = data
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
