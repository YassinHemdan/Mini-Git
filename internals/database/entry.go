package internals

type Entry interface {
	// Object
	GetName() string
	GetMode() string
	GetPathname() string
	ParentDirectories() []string
	Type() string
	GetOid() []byte
}
