package internals

type Entry interface {
	GetName() string
	GetMode() string
	Type() string
	GetOid() []byte
}
