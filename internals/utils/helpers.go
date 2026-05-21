package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var invalidName = regexp.MustCompile(`^\.|/\.|\.\.|^/|/$|\.lock$|@\{|[\x00-\x20*:?\[\\^~\x7f]`)

/*
Returns all parent directories for a given pathname
let pathname = "a/b/c/file.txt"

parentDirectores = ["a", "a/b", "a/b/c"]

notice that a/b/c/file.txt is not involved, we are only intereseted in the parents dirs
*/
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
