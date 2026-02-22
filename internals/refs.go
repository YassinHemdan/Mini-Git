package internals

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

type Refs struct {
	path_name string // the jit path. I don't like it
}

func (r *Refs) New(path_name string) error {
	r.path_name = path_name

	return nil
}

func (r *Refs) UpdateHead(oid []byte) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	file, err := os.OpenFile(r.headPath(), flags, JitDefaultPermission)

	if err != nil {
		return fmt.Errorf("Error: Couldn't write to HEAD file - %v", err)
	}

	if _, err := fmt.Fprintf(file, "%x", oid); err != nil {
		return fmt.Errorf("Error: Couldn't write to HEAD file - %v", err)
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
	return hex.DecodeString(file_content.String())
}

func (r *Refs) headPath() string {
	return strings.Join([]string{r.path_name, "HEAD"}, string(os.PathSeparator))
}
