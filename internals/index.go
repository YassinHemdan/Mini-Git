package internals

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"slices"
	"syscall"
)

type entry struct {
	ctime     uint32
	ctimeNsec uint32
	mtime     uint32
	mtimeNsec uint32
	dev       uint32
	ino       uint32
	mode      uint32
	uid       uint32
	gid       uint32
	size      uint32
	oid       []byte
	flags     uint16
	path      string
}

func newEntry(pathname string, oid []byte, stat *syscall.Stat_t) (*entry, error) { // what do we need for every entry ? pathname ?

	namelen := uint32(min(0xFFF, len(pathname)))
	var mode uint32 = 0100644

	if stat.Mode&syscall.S_IFREG != 0 && stat.Mode&0111 != 0 {
		mode = 0100755
	}

	return &entry{
		ctime:     uint32(stat.Ctim.Sec),
		ctimeNsec: uint32(stat.Ctim.Nsec),
		mtime:     uint32(stat.Mtim.Sec),
		mtimeNsec: uint32(stat.Mtim.Nsec),
		dev:       uint32(stat.Dev),
		ino:       uint32(stat.Ino),
		mode:      mode,
		uid:       stat.Uid,
		gid:       stat.Gid,
		size:      uint32(stat.Size),
		oid:       oid,
		flags:     uint16(namelen),
		path:      pathname,
	}, nil
}

func (e *entry) toBytes() []byte {
	buffer := new(bytes.Buffer)

	// the order matters
	binary.Write(buffer, binary.BigEndian, e.ctime)
	binary.Write(buffer, binary.BigEndian, e.ctimeNsec)
	binary.Write(buffer, binary.BigEndian, e.mtime)
	binary.Write(buffer, binary.BigEndian, e.mtimeNsec)
	binary.Write(buffer, binary.BigEndian, e.dev)
	binary.Write(buffer, binary.BigEndian, e.ino)
	binary.Write(buffer, binary.BigEndian, e.mode)
	binary.Write(buffer, binary.BigEndian, e.uid)
	binary.Write(buffer, binary.BigEndian, e.gid)
	binary.Write(buffer, binary.BigEndian, e.size)
	binary.Write(buffer, binary.BigEndian, e.oid)
	binary.Write(buffer, binary.BigEndian, e.flags)
	buffer.WriteString(e.path)
	buffer.WriteByte(0)

	for buffer.Len()%8 != 0 {
		buffer.WriteByte(0)
	}

	return buffer.Bytes()
}

func (e *entry) key() string {
	return e.path
}

type Index struct {
	keys     []string
	entries  map[string]*entry
	lockfile *LockFile
}

func NewIndex(pathname string) (*Index, error) {
	lf := LockFile{}
	if err := lf.New(pathname); err != nil {
		return nil, fmt.Errorf("Could not create an index file - %v", err)
	}

	return &Index{
		entries:  make(map[string]*entry),
		lockfile: &lf,
	}, nil
}

func (idx *Index) Add(pathname string, oid []byte, stat *syscall.Stat_t) error {
	entry, err := newEntry(pathname, oid, stat)
	if err != nil {
		return fmt.Errorf("Can't create index entry - %v", err)
	}
	idx.keys = append(idx.keys, entry.key())
	idx.entries[entry.key()] = entry
	return nil
}

func (idx *Index) WriteUpdates() error {
	if isLocked, err := idx.lockfile.HoldForUpdate(); !isLocked || err != nil {
		return fmt.Errorf("Could not acquire lock over index file - %v", err)
	}
	fmt.Println("Hi")
	var digest hash.Hash = sha1.New()

	header := new(bytes.Buffer)
	header.WriteString("DIRC")
	binary.Write(header, binary.BigEndian, uint32(2))
	binary.Write(header, binary.BigEndian, uint32(len(idx.entries)))

	if err := idx.write(header.Bytes(), digest); err != nil {
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
		if err := idx.write(entry.toBytes(), digest); err != nil {
			return fmt.Errorf("Could not write in index file - %v", err)
		}
	}

	if err := idx.finishWrite(digest); err != nil {
		return fmt.Errorf("Could not finish writing in index file - %v", err)
	}

	return nil
}

func (idx *Index) write(data []byte, digest hash.Hash) error {
	if err := idx.lockfile.Write(string(data)); err != nil {
		return fmt.Errorf("Could not write in index file - %v", err)
	}
	if _, err := digest.Write(data); err != nil {
		return fmt.Errorf("Could not digest index file's data - %v", err)
	}

	return nil
}
func (idx *Index) finishWrite(digest hash.Hash) error {

	hash := digest.Sum(nil)

	if err := idx.lockfile.Write(string(hash)); err != nil {
		return fmt.Errorf("Could not write in index file - %v", err)
	}

	if err := idx.lockfile.Save(); err != nil {
		return fmt.Errorf("Could not commit index/locked file - %v", err)
	}
	return nil
}
