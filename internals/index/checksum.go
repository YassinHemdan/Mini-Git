package index

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"os"
)

const (
	CHECKSUM_SIZE = 20
)

type checksum struct {
	digest hash.Hash
	file   *os.File
}

func newChecksum(file *os.File) *checksum {
	return &checksum{
		digest: sha1.New(),
		file:   file,
	}
}
func (c *checksum) read(size int) ([]byte, error) {
	data := make([]byte, size)
	if _, err := io.ReadFull(c.file, data); err != nil {
		return nil, fmt.Errorf("Can't read index file with size %d - %v", size, err)
	}

	c.digest.Write(data)

	return data, nil
}

func (c *checksum) validateChecksum() (bool, error) {
	data := make([]byte, CHECKSUM_SIZE)
	if _, err := io.ReadFull(c.file, data); err != nil {
		return false, fmt.Errorf("Can't validate checksum - %v", err)
	}

	return string(data) == string(c.digest.Sum(nil)), nil
}

func (c *checksum) write(data []byte) error {
	if _, err := c.file.Write(data); err != nil {
		return fmt.Errorf("Could not write in index file - %v", err)
	}

	if _, err := c.digest.Write(data); err != nil {
		return fmt.Errorf("Could not digest data - %v", err)
	}
	return nil
}

func (c *checksum) writeChecksum() error {
	data := c.digest.Sum(nil)
	if _, err := c.file.Write(data); err != nil {
		return fmt.Errorf("Could not write in index file - %v", err)
	}
	return nil
}
