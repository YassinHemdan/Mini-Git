package index

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	ENTRY_MIN_SIZE = 64
	ENTRY_BLOCK    = 8
)

// we need to make this entry implement the Entry interface
type Entry struct {
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

func newEntry(pathname string, oid []byte, stat *syscall.Stat_t) (*Entry, error) {

	namelen := uint32(min(0xFFF, len(pathname)))
	var mode uint32 = 0100644

	if stat.Mode&syscall.S_IFREG != 0 && stat.Mode&0111 != 0 {
		mode = 0100755
	}

	return &Entry{
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

func ParseEntry(data []byte) (*Entry, error) {
	if len(data) < ENTRY_MIN_SIZE {
		return nil, fmt.Errorf("entry data too short: %d bytes", len(data))
	}
	ctime := binary.BigEndian.Uint32(data[0:4])
	ctimeNsec := binary.BigEndian.Uint32(data[4:8])
	mtime := binary.BigEndian.Uint32(data[8:12])
	mtimeNsec := binary.BigEndian.Uint32(data[12:16])
	dev := binary.BigEndian.Uint32(data[16:20])
	ino := binary.BigEndian.Uint32(data[20:24])
	mode := binary.BigEndian.Uint32(data[24:28])
	uid := binary.BigEndian.Uint32(data[28:32])
	gid := binary.BigEndian.Uint32(data[32:36])
	size := binary.BigEndian.Uint32(data[36:40])

	oid := make([]byte, 20)
	copy(oid, data[40:60])

	flags := binary.BigEndian.Uint16(data[60:62])

	// the min length of the whole entry will be 64 byte (0 - 63)
	/*
		the file's name will be between flags and the null terminator
		so we will locate the null's location to be able to get the file's name

		file's name [62: nullIdx]
	*/
	pathBytes := data[62:]

	nullIdx := -1
	for i, b := range pathBytes {
		if b == 0 {
			nullIdx = i
			break
		}
	}
	if nullIdx == -1 {
		return nil, fmt.Errorf("no null terminator found in entry path")
	}
	path := string(pathBytes[:nullIdx]) // file name directly before the null terminator

	return &Entry{
		ctime:     ctime,
		ctimeNsec: ctimeNsec,
		mtime:     mtime,
		mtimeNsec: mtimeNsec,
		dev:       dev,
		ino:       ino,
		mode:      mode,
		uid:       uid,
		gid:       gid,
		size:      size,
		oid:       oid,
		flags:     flags,
		path:      path,
	}, nil
}
func (e *Entry) toBytes() []byte {
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
func (e *Entry) key() string {
	return e.path
}
func (e *Entry) GetName() string {
	return filepath.Base(e.path)
}
func (e *Entry) GetMode() string {
	return fmt.Sprintf("%o", e.mode)
}
func (e *Entry) GetPathname() string {
	return e.path
}
func (e *Entry) GetOid() []byte {
	return e.oid
}
func (e *Entry) ParentDirectories() []string {
	prefixs := strings.Split(filepath.ToSlash(e.GetPathname()), "/")
	parents := []string{}

	for i := 1; i < len(prefixs); i++ {
		parents = append(parents, strings.Join(prefixs[:i], "/"))
	}

	return parents
}
func (e *Entry) Type() string {
	return "IndexEntry"
}
