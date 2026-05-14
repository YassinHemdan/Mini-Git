package internals

import (
	"JIT/internals/locks"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var invalidName = regexp.MustCompile(`^\.|/\.|\.\.|^/|/$|\.lock$|@\{|[\x00-\x20*:?\[\\^~\x7f]`)

const (
	HEAD = "HEAD"
)

type InvalidBranch struct {
	message string
}

func (e *InvalidBranch) Error() string {
	return fmt.Sprintf("fatal: %s", e.message)
}

type lockDenied struct {
	message string
}

func (l *lockDenied) Error() string {
	return fmt.Sprintf("lock request has beed denied: %s", l.message)
}

type Refs struct {
	pathname  string
	refsPath  string
	headsPath string
}

func NewRefs(pathname string) (*Refs, error) {
	refsPath := filepath.Join(pathname, "refs")
	headsPath := filepath.Join(refsPath, "heads")

	return &Refs{
		pathname:  pathname,
		refsPath:  refsPath,
		headsPath: headsPath,
	}, nil
}

func (r *Refs) CreateBranch(branchName string) error {
	/*
		to create a branch, we wanna make sure its name is valid and it was not created before
	*/
	if invalidName.MatchString(branchName) {
		return &InvalidBranch{
			message: fmt.Sprintf("fatal: '%s' is not a valid branch name", branchName),
		}
	}

	branchPath := filepath.Join(r.headsPath, branchName)
	if _, err := os.Stat(branchPath); err == nil {
		return &InvalidBranch{
			message: fmt.Sprintf("fatal: A branch named '%s' already exists", branchName),
		}
	}

	oid, err := r.ReadHead()
	if err != nil {
		return err
	}

	return r.updateRefFile(branchPath, oid)
}

func (r *Refs) updateRefFile(path string, oid []byte) error {
	lockfile := locks.LockFile{}
	if err := lockfile.New(path); err != nil {
		return fmt.Errorf("Error: Couldn't make a new lockfile - %v", err)
	}

	success, err := lockfile.HoldForUpdate()
	if err != nil {
		return err
	}

	if !success {
		return &lockDenied{message: "Could not acquire lock on file: " + path}
	}

	if err := lockfile.Write(fmt.Sprintf("%x\n", oid)); err != nil {
		return fmt.Errorf("Error: Couldn't make write to lockfile - %v", err)
	}

	if err := lockfile.Save(); err != nil {
		return fmt.Errorf("Error: Couldn't make save the lockfile - %v", err)
	}

	return nil
}
func (r *Refs) UpdateHead(data []byte) error {
	return r.updateRefFile(r.getHeadPath(), data)
}
func (r *Refs) ReadHead() ([]byte, error) {
	file, err := os.Open(r.getHeadPath())

	if err != nil {
		return nil, nil // file does not exist
	}

	var file_content bytes.Buffer

	if _, err := io.Copy(&file_content, file); err != nil {
		return nil, fmt.Errorf("Couldn't copy from HEAD file - %v", err)
	}

	file.Close()

	return hex.DecodeString(strings.TrimSpace(file_content.String()))
}

func (r *Refs) getHeadPath() string {
	return filepath.Join(r.pathname, HEAD)
}
