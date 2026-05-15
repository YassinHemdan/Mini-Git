package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var invalidName = regexp.MustCompile(`^\.|/\.|\.\.|^/|/$|\.lock$|@\{|[\x00-\x20*:?\[\\^~\x7f]`)

func ParentDirectories(pathname string) []string {
	prefixs := strings.Split(filepath.ToSlash(pathname), "/")
	parents := []string{}

	for i := 1; i < len(prefixs); i++ {
		parents = append(parents, strings.Join(prefixs[:i], "/"))
	}

	return parents
}

func IsValidName(name string) bool {
	return !invalidName.MatchString(name)
}
