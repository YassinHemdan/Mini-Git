package utils

import (
	"path/filepath"
	"strings"
)

func ParentDirectories(pathname string) []string {
	prefixs := strings.Split(filepath.ToSlash(pathname), "/")
	parents := []string{}

	for i := 1; i < len(prefixs); i++ {
		parents = append(parents, strings.Join(prefixs[:i], "/"))
	}

	return parents
}
