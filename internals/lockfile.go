package internals

import (
	"fmt"
	"os"
)

type LockFile struct {
	file_path string
	lock_path string
	lockfile  *os.File
}

func (l *LockFile) New(file_path string) error {
	l.file_path = file_path
	l.lock_path = file_path + ".lock"
	l.lockfile = nil

	return nil
}
func (l *LockFile) HoldForUpdate() (bool, error) {
	if l.lockfile == nil {
		fmt.Println("I am here - 1")
		flags := os.O_CREATE | os.O_EXCL | os.O_RDWR
		lock_file, err := os.OpenFile(l.lock_path, flags, JitDefaultPermission)

		if err != nil {
			if os.IsExist(err) {
				fmt.Println("I am here - 2")
				return false, nil
			}
			if os.IsNotExist(err) {
				fmt.Println("I am here - 3")
				return false, &MissingParent{message: "Parent Dir Not Found"}
			}
			if os.IsPermission(err) {
				fmt.Println("I am here - 4")
				return false, &NoPermission{message: "Invalid Permissions"}
			}
			fmt.Println("I am here - 5")
			return false, err
		}

		l.lockfile = lock_file
	}
	return true, nil
}
func (l *LockFile) Write(data string) error {
	if !l.isLocked() {
		return &StaleLock{message: "A lock is required"}
	}

	if _, err := l.lockfile.WriteString(data); err != nil {
		return fmt.Errorf("Error: Could not write to lock file - %v", err)
	}
	return nil
}
func (l *LockFile) Save() error {
	if !l.isLocked() {
		return &StaleLock{message: "A lock is required"}
	}

	l.lockfile.Close()

	if err := os.Rename(l.lock_path, l.file_path); err != nil {
		return fmt.Errorf("failed to rename the new file - %v", err)
	}

	l.lockfile = nil

	return nil
}

func (l *LockFile) isLocked() bool {
	return l.lockfile != nil
}
