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
	path_name string // the jit path. I don't like it
}

func (r *Refs) New(path_name string) error {
	r.path_name = path_name

	return nil
}

func (r *Refs) UpdateHead(oid []byte) error {
	lock_file := LockFile{}
	if err := lock_file.New(r.headPath()); err != nil {
		return fmt.Errorf("Error: Couldn't make a new lockfile - %v", err)
	}

	success, err := lock_file.HoldForUpdate()
	if err != nil {
		return err
	}
	if !success {
		return &lockDenied{message: "Could not acquire lock on file: " + r.headPath()}
	}

	if err := lock_file.Write(fmt.Sprintf("%x\n", oid)); err != nil {
		return fmt.Errorf("Error: Couldn't make write to lockfile - %v", err)
	}

	if err := lock_file.Save(); err != nil {
		return fmt.Errorf("Error: Couldn't make save the lockfile - %v", err)
	}
	return nil
}

func (r *Refs) ReadHead() ([]byte, error) {
	file, err := os.Open(r.headPath())
	if err != nil {
		// means file does not exist
		return nil, nil
	}

	var file_content bytes.Buffer

	if _, err := io.Copy(&file_content, file); err != nil {
		return nil, fmt.Errorf("Couldn't copy from HEAD file - %v", err)
	}

	file.Close()

	// str_file_content := file_content.String()[:]
	return hex.DecodeString(strings.TrimSpace(file_content.String()))
}

func (r *Refs) headPath() string {
	return strings.Join([]string{r.path_name, "HEAD"}, string(os.PathSeparator))
}
