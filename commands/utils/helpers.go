package utils

import (
	"JIT/internals"
	"JIT/internals/utils"
	"path/filepath"
)

func Repo(dir string) (*internals.Repository, error) {
	jit_dir := filepath.Join(dir, utils.JitMetadataDir)
	return internals.NewRepository(jit_dir)
}
