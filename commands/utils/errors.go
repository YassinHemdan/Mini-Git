package utils

import "fmt"

type MissingFileError struct {
	Path string
	Err  error
}

func (e *MissingFileError) Error() string {
	return fmt.Sprintf("pathspec '%s' did not match any files", e.Path)
}

func (e *MissingFileError) Unwrap() error {
	return e.Err
}

type NoPermissionError struct {
	Path string
	Err  error
}

func (e *NoPermissionError) Error() string {
	return fmt.Sprintf("open('%s'): Permission denied", e.Path)
}

func (e *NoPermissionError) Unwrap() error {
	return e.Err
}
