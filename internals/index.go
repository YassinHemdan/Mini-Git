package internals

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"syscall"
)

/*
	Lets see how the index file is formatted
	FIRST: 12-byte of header consisting of:
		1- 4-byte signature: 'D' 'I' 'R' 'C' stands for dircache
		2- 4-byte version number -> it will be version 1 ==> 1
		3- 32-bit number of index entries == 4-byte

		it will similar to this:


		00000000 44 49 52 43 00 00 00 02 00 00 00 01       |DIRC........ |



		44 49 52 43 => D I R C
		00 00 00 02 => 2
		00 00 00 01 => 1       we will only add one entry for now to our index


	SECOND: The header is followed by the entries themselves,
		FOR NOW: we will append only one entry, lets see how the entry will be formatted
		ENTRY: each entry begins with some values that we can get by calling stat() on the file,
		   	comprising 10 4-byte numbers in all

		   		1-  32-bit ctime seconds   =>the last time a file's "metadata" changed  (change time)
		   		2-  32-bit ctime nanoseconds fractions
		   		3-  32-bit mtime seconds   => the last time a file's "data" changed  (modify time)
		   		4-  32-bit mtime nanoseconds fractions
		   		5-  32-bit dev  => the ID of the hardware device the file resides on
		   		6-  32-bit ino  => the inode storing attributes
		   		7-  32-bit mode => file modes like before
		   		8-  32-bit uid  => ID of the file's user
		   		9-  32-bit gid  => ID of the group
		   		10- 32 bit file size

	THIRD:  160-bit (20-byte) SHA-1 of the object. This is the blob's id created in the .git/objects
	FOURTH: 16-bit (2-byte) set of flags
	FIFTH:  12-bit file's name length if the length is less than 0xFFF;
			otherwise 0xFFF is stored in this field
	SIXTH: another 20-byte SHA-1 hash, it is the hash if the index itself
*/

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
	// binary.Write(buffer, binary.BigEndian, e.path)
	buffer.WriteString(e.path)
	buffer.WriteByte(0)

	for buffer.Len()%8 != 0 {
		buffer.WriteByte(0)
	}

	return buffer.Bytes()
}

type Index struct {
	// what do we need here ?
	entries  map[string]*entry // shall we make it a normal slice instead of a map ?
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

	idx.entries[pathname] = entry
	return nil
}

func (idx *Index) WriteUpdates() error {

	/*
		we will use our lockfile to make sure that we are the only ones modifying it
		1-  Make sure that we aquired the lock over our file
		2- Prepare the header format, save it and digest it
		3- loop over entries, get the toString for every one and save it, digest it
	*/

	/*
		In the future, we need to make sure that this is the right way to handle the failure of
		HoldForUpdate
	*/
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

	for _, entry := range idx.entries {
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
