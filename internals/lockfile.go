// I will create a lockfile class that maintains the updating for the HEAD file
/*
	We wanna make sure that there will be no conflicts when two or more processes try to write on the HEAD file
	as we are not sure that these two processes will write the same data or not
	so we make each action to be done "Atomic"
*/

package internals

import (
	"fmt"
	"os"
)

type LockFile struct {
	// HEAD file path
	// lock file path
	// actual lock file

	file_path string // head
	lock_path string
	lockfile  *os.File
}

func (l *LockFile) New(file_path string) error {
	l.file_path = file_path
	l.lock_path = file_path + ".lock"
	l.lockfile = nil

	return nil
}

/*
	we will create a function called HoldForUpdate() (bool, error)
	(true, nil) => request was successful
	(false, nil) => request was denied
	(false, error) => an error occured

	it will lock our file to make sure that no one will be able to write on this file when we are working on it

	1- if the file does not exist, the HEAD file in our case, it means that it was not created in the first
		place and we can't perform commits, so we will return (false, error)
	2- if it exists, we will check for the .lock file
	  a- if the .lock file does not exist, it means no one is working on the HEAD file and we will create
	  	it for our selves. that means our request to acquire a lock is successful. return (true, nil)
	  b- if the .lock file exists, someone is working on it and our request was denied
	  		return (fasle, nil)
	  c- if the parentDir not exists or the premissions were invalied, return (fasle, error)

	  ### notice that in 2.b we did not return an error as its just a denied request and we can try again later

*/

func (l *LockFile) HoldForUpdate() (bool, error) {
	if l.lockfile == nil {
		flags := os.O_CREATE | os.O_EXCL | os.O_RDWR
		lock_file, err := os.OpenFile(l.lock_path, flags, JitDefaultPermission)

		if err != nil {
			if os.IsExist(err) {
				return false, nil
			}
			if os.IsNotExist(err) {
				return false, &MissingParent{message: "Parent Dir Not Found"}
			}
			if os.IsPermission(err) {
				return false, &NoPermission{message: "Invalid Permissions"}
			}

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