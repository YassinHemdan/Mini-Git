package internals

type Object interface {
	ToString() string
	GetOid() []byte
	SetOid([]byte)
	Type() string
}
