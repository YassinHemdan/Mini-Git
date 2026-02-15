package internals

import "os"

const (
	JitMetadataDir       = ".jit"
	JitDefaultPermission = os.FileMode(0744)
	JitMetadataContent   = "objects|refs"
)
