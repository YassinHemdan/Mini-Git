package internals

import "fmt"

type MissingParent struct {
	message string
}

type NoPermission struct {
	message string
}

type StaleLock struct {
	message string
}

func (e *MissingParent) Error() string {
	return fmt.Sprintf("missing parent: %s", e.message)
}

func (e *NoPermission) Error() string {
	return fmt.Sprintf("file does not exist: %s", e.message)
}

func (e *StaleLock) Error() string {
	return fmt.Sprintf("lock is required: %s", e.message)
}
