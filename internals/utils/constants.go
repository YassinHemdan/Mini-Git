package utils

import "os"

const (
	JitMetadataDir       = ".jit"
	JitDefaultPermission = os.FileMode(0744)
	JitMetadataContent   = "objects|refs"
)
