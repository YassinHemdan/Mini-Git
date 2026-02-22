package internals

import (
	"fmt"
	"os"
)

type LockFile struct {
	file_path string
	lock_path string
	lock_file *os.File
}

func (l *LockFile) New(file_path string) error {
	l.file_path = file_path
	l.lock_path = file_path + ".lock"
	l.lock_file = nil

	return nil
}

func (l *LockFile) HoldForUpdate() (bool, error) {
	if l.lock_file == nil {
		// we need to create a lock file if it is not existing
		// if it exists, catch the error and return false
		flags := os.O_RDWR | os.O_EXCL | os.O_CREATE
		lock_file, err := os.OpenFile(l.lock_path, flags, JitDefaultPermission)
		if err != nil {
			if os.IsExist(err) {
				return false, nil
			}
			if os.IsNotExist(err) {
				return false, &MissingParent{message: "Parent dir does not exist"}
			}
			if os.IsPermission(err) {
				return false, &NoPermission{message: "Invalid Permissions"}
			}

			return false, err
		}
		l.lock_file = lock_file
	}
	return true, nil
}

func (l *LockFile) Write(content string) error {
	if !l.isLocked() {
		return &StaleLock{message: "A lock is required"}
	}

	if _, err := l.lock_file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to lock file - %v", err)
	}

	return nil
}

func (l *LockFile) Save() error {
	if !l.isLocked() {
		return &StaleLock{message: "A lock is required"}
	}

	if err := l.lock_file.Close(); err != nil {
		return fmt.Errorf("failed to close lock file - %v", err)
	}

	if err := os.Rename(l.lock_path, l.file_path); err != nil {
		return fmt.Errorf("failed to rename the new file - %v", err)
	}

	l.lock_file = nil

	return nil

}

func (l *LockFile) isLocked() bool {
	return l.lock_file != nil
}
