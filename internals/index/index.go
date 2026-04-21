package index

import (
	"JIT/internals/locks"
	"JIT/internals/utils"
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
	keys      map[string]bool
	entries   map[string]*Entry
	parents   map[string]map[string]bool
	lockfile  *locks.LockFile
	pathname  string
	isChanged bool
}

func NewIndex(pathname string) (*Index, error) {
	lf := locks.LockFile{}
	if err := lf.New(pathname); err != nil {
		return nil, fmt.Errorf("Could not create an index file - %v", err)
	}

	return &Index{
		entries:   make(map[string]*Entry),
		keys:      make(map[string]bool),
		parents:   make(map[string]map[string]bool),
		lockfile:  &lf,
		pathname:  pathname,
		isChanged: false,
	}, nil
}

func (idx *Index) Add(pathname string, oid []byte, stat *syscall.Stat_t) error {
	if len(pathname) == 0 {
		return fmt.Errorf("Can't add an entry with no name")
	}
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
	sortedKeys := idx.getKeysSlice()
	for _, key := range sortedKeys {
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
	sortedKeys := idx.getKeysSlice()

	for _, key := range sortedKeys {
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
	idx.resolveConflicts(entry)
	val, exists := idx.entries[entry.key()]
	if !exists {
		// first time to save it
		idx.entries[entry.key()] = entry
		idx.keys[entry.key()] = true
		parentDirs := entry.ParentDirectories()
		for _, parentname := range parentDirs {
			if _, ok := idx.parents[parentname]; !ok {
				idx.parents[parentname] = make(map[string]bool)
			}
			idx.parents[parentname][entry.key()] = true
		}
	} else if string(val.GetOid()) != string(entry.GetOid()) {
		// modified
		idx.entries[entry.key()] = entry
	}

	return nil
}

func (idx *Index) resolveConflicts(entry *Entry) { // logl + n*l
	idx.replacingFileWithDirectoryCheck(entry)
	idx.replacingDirectoryWithFile(entry)
}

func (idx *Index) replacingFileWithDirectoryCheck(entry *Entry) { // O(L + LlogL), where n is the length file's ParentDirectories
	/*
		--> we will take the path and we will check if any of the parents of the new path
		is located in our keys or not, if so, it means there was a file with the same name of a
		parent directory of the given path, so we will remove it
		... we can binary search in the ParentDirectories of the path
	*/

	check := func(target string) int {
		_, inParents := idx.parents[target]
		_, inKeys := idx.keys[target]

		if inParents && !inKeys {
			return 1
		} else if !inParents && !inKeys {
			return -1
		}
		return 0
	}

	parentDirectories := entry.ParentDirectories()

	s, e := 0, len(parentDirectories)-1

	ans := -1
	for s <= e {
		mid := s + (e-s)/2
		state := check(parentDirectories[mid])
		if state == 0 {
			ans = mid
			break
		} else if state == 1 {
			s = mid + 1
		} else {
			e = mid - 1
		}
	}

	if ans == -1 { // No answer found, we don't need to remove anything
		return
	}

	pathnameToRemove := parentDirectories[ans]
	idx.removeEntry(pathnameToRemove) // O(L)

}

func (idx *Index) replacingDirectoryWithFile(entry *Entry) {
	/*
		O(L * N):
			where N is the length if the innerMap (#files under the current directory)
			where L is the length if file's ParentDirectories
	*/
	/*
		we have a parents map:
		key1 => parentPathname
		val1 => map:  key2 => filepathname, val2 => bool "always true"

		so, if the filepath exists as key1, that means it was a dir before
		we will remove it and also we will loop over key2 (all files under it) and remove them from our entries
	*/

	pathname := entry.GetPathname()

	if innerMap, ok := idx.parents[pathname]; ok {
		for filename := range innerMap { // O(L * N)
			idx.removeEntry(filename) // O(L)
		}
	}

	delete(idx.parents, pathname) // don't forget to remove the parent (directory)
}

func (idx *Index) removeEntry(pathname string) { // O(L) where n is the length of file's ParentDirectories
	pathParents := utils.ParentDirectories(pathname)
	for _, curParent := range pathParents {
		childsMap := idx.parents[curParent]
		delete(childsMap, pathname) // delete the child
		if len(childsMap) == 0 {    // if the map of childs got empty.. delete the parent
			delete(idx.parents, curParent)
		}
	}
	delete(idx.keys, pathname)
	delete(idx.entries, pathname)
}

func (idx *Index) Clear() {
	idx.entries = make(map[string]*Entry)
	idx.parents = make(map[string]map[string]bool)
	idx.keys = make(map[string]bool, 0)
	idx.isChanged = false
}

func (idx *Index) getKeysSlice() []string {
	result := make([]string, 0)
	for k := range idx.keys {
		result = append(result, k)
	}

	slices.Sort(result)

	return result
}

func (idx *Index) ReleaseLock() error {
	if err := idx.lockfile.Rollback(); err != nil {
		return fmt.Errorf("Could not release index lock - %v", err)
	}

	return nil
}

func (idx *Index) IsTracked(pathname string) bool {
	_, ok1 := idx.keys[pathname]
	_, ok2 := idx.parents[pathname] // a directory is tracked, no need to check for its childs
	return ok1 || ok2
}

func (idx *Index) UpdateEntryStat(entry *Entry, stat *syscall.Stat_t) {
	entry.updateState(stat)
	idx.isChanged = true
}
