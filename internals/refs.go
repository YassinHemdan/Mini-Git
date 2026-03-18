package internals

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

type lockDenied struct {
	message string
}

func (l *lockDenied) Error() string {
	return fmt.Sprintf("lock request has beed denied: %s", l.message)
}

type Refs struct {
	path_name string
}

func (r *Refs) New(path_name string) error {
	r.path_name = path_name // this will be the /.jit

	return nil
}

func (r *Refs) UpdateHead(data []byte) error {
	lockfile := LockFile{}

	if err := lockfile.New(r.getHeadPath()); err != nil {
		return fmt.Errorf("Error: Couldn't make a new lockfile - %v", err)
	}

	success, err := lockfile.HoldForUpdate()

	if err != nil {
		return err
	}
	if !success {
		return &lockDenied{message: "Could not acquire lock on file: " + r.getHeadPath()}
	}
	if err := lockfile.Write(fmt.Sprintf("%x\n", data)); err != nil {
		return fmt.Errorf("Error: Couldn't make write to lockfile - %v", err)
	}
	if err := lockfile.Save(); err != nil {
		return fmt.Errorf("Error: Couldn't make save the lockfile - %v", err)
	}

	return nil
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
	return strings.Join([]string{r.path_name, "HEAD"}, string(os.PathSeparator))
}