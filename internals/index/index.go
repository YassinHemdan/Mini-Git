package index

import (
	"JIT/internals/locks"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"syscall"
)

const (
	SIGNATURE   = "DIRC"
	VERSION     = 2
	HEADER_SIZE = 12
)

type Index struct {
	keys      []string
	entries   map[string]*Entry
	lockfile  *locks.LockFile
	pathname  string
	isChanged bool
}

func NewIndex(pathname string) (*Index, error) {
	// fmt.Println("Index name = ", pathname)
	lf := locks.LockFile{}
	if err := lf.New(pathname); err != nil {
		return nil, fmt.Errorf("Could not create an index file - %v", err)
	}

	return &Index{
		entries:   make(map[string]*Entry),
		lockfile:  &lf,
		pathname:  pathname,
		isChanged: false,
	}, nil
}

func (idx *Index) Add(pathname string, oid []byte, stat *syscall.Stat_t) error {

	entry, err := newEntry(pathname, oid, stat)
	if err != nil {
		return fmt.Errorf("Can't create index entry - %v", err)
	}

	if err := idx.storeEntry(entry); err != nil {
		return fmt.Errorf("Can't store index entry - %v", err)
	}

	idx.isChanged = true
	return nil
}

func (idx *Index) GetEntries() []*Entry {
	entries := make([]*Entry, 0)
	for _, key := range idx.keys {
		entry := idx.entries[key]
		entries = append(entries, entry)
	}

	return entries
}
func (idx *Index) WriteUpdates() error {
	// we will not acquire a lock here, it should be already acquired when reading it

	if !idx.isChanged {
		return idx.lockfile.Rollback()
	}

	writer := newChecksum(idx.lockfile.GetLockfile())

	header := new(bytes.Buffer)
	header.WriteString(SIGNATURE)
	binary.Write(header, binary.BigEndian, uint32(VERSION))
	binary.Write(header, binary.BigEndian, uint32(len(idx.entries)))

	if err := writer.write(header.Bytes()); err != nil {
		return fmt.Errorf("Could not write in index file - %v", err)
	}

	slices.SortFunc(idx.keys, func(str1, str2 string) int {
		if str1 < str2 {
			return -1
		} else if str1 > str2 {
			return 1
		} else {
			return 0
		}
	})

	for _, key := range idx.keys {
		entry := idx.entries[key]

		if err := writer.write(entry.toBytes()); err != nil {
			return fmt.Errorf("Could not write in index file - %v", err)
		}
	}
	if err := writer.writeChecksum(); err != nil {
		return fmt.Errorf("Could not finish writing in index file - %v", err)
	}

	// free the lock
	if err := idx.lockfile.Save(); err != nil {
		return fmt.Errorf("Could not commit changes on index file - %v", err)
	}

	idx.isChanged = false

	return nil
}
func (idx *Index) LoadForUpdate() (bool, error) {
	if isLocked, err := idx.lockfile.HoldForUpdate(); !isLocked || err != nil {
		return false, fmt.Errorf("Couldn't acquire lock on index file - %v", err)
	}

	verified, err := idx.Load()
	if err != nil {
		return false, err
	}

	return verified, nil
}
func (idx *Index) Load() (bool, error) {
	idx.Clear()
	file, err := idx.openIndexFile()

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		return false, err
	}

	defer file.Close()

	reader := newChecksum(file)
	count, err := idx.readHeader(reader)
	if err != nil {
		return false, fmt.Errorf("Could not load index content - %v", err)
	}

	if err := idx.readEntries(reader, count); err != nil {
		return false, fmt.Errorf("Could not load index content - %v", err)
	}

	return reader.validateChecksum()
}

func (idx *Index) openIndexFile() (*os.File, error) {
	file, err := os.Open(idx.pathname)
	if err != nil {
		return nil, fmt.Errorf("could not open index file - %w", err)
	}
	return file, nil
}
func (idx *Index) readHeader(reader *checksum) (uint32, error) {
	data, err := reader.read(HEADER_SIZE)
	if err != nil {
		return 0, fmt.Errorf("Couldn't read index's header - %v", err)
	}

	// translate the data to signature, version and count

	signature := string(data[:4])
	version := binary.BigEndian.Uint32(data[4:8])
	count := binary.BigEndian.Uint32(data[8:12])

	if signature != SIGNATURE {
		return 0, fmt.Errorf("Signature: expected '%s' but found '%s'", SIGNATURE, signature)
	}

	if version != VERSION {
		return 0, fmt.Errorf("Version: expected '%d' but found '%d'", VERSION, version)
	}

	return count, nil
}

func (idx *Index) readEntries(reader *checksum, count uint32) error {
	for count != 0 {
		data, err := reader.read(ENTRY_MIN_SIZE) // [64] byte
		if err != nil {
			return fmt.Errorf("Error: Could not read entry min size - %v", err)
		}

		// if the last byte is not a 00, we will read the next entry block (8 bytes)

		for data[len(data)-1] != 0 {
			block, err := reader.read(ENTRY_BLOCK)
			if err != nil {
				return fmt.Errorf("Error: Could not entry block - %v", err)
			}

			data = append(data, block...)
		}

		// now we are sure that we got the entry with tis padding zeros that makes sure it is a
		// mutiple of 8

		// now we will parse it using our entry struct and it will return for us *entry
		entry, err := ParseEntry(data)

		if err != nil {
			return fmt.Errorf("Error: Could not parse entry - %v", err)
		}
		if err := idx.storeEntry(entry); err != nil {
			return fmt.Errorf("Error: Could not store entry - %v", err)
		}

		count--
	}

	return nil
}

func (idx *Index) storeEntry(entry *Entry) error {
	idx.entries[entry.key()] = entry
	idx.keys = append(idx.keys, entry.key())

	return nil
}
func (idx *Index) Clear() {
	idx.entries = make(map[string]*Entry)
	idx.keys = make([]string, 0)
	idx.isChanged = false
}
