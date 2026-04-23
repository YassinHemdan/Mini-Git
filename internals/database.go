package internals

import (
	database "JIT/internals/database"
	"JIT/internals/utils"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

type IDatabase interface {
	New(string) error
	Store(database.Object) error
}

type Database struct {
	count int
	path  string
}

func NewDatabase(pathname string) (*Database, error) {
	return &Database{
		count: 0,
		path:  pathname,
	}, nil
}

func (db *Database) Store(object database.Object) error {
	content, err := db.serializeObject(object)
	if err != nil {
		return err
	}
	object.SetOid(db.hashContent(content))
	return db.writeObject(object.GetOid(), content)
}

func (db *Database) writeObject(oid, data []byte) error {
	// oid_hex := fmt.Sprintf("%x", oid)
	// object_dir := strings.Join([]string{db.path, oid_hex[:2]}, string(os.PathSeparator))
	// object_path := strings.Join([]string{object_dir, oid_hex[2:]}, string(os.PathSeparator))
	object_path := db.objectPath(oid)
	object_dir := filepath.Dir(object_path)
	if _, err := os.Stat(object_path); err == nil {
		return nil
	}

	db.count++

	if err := os.MkdirAll(object_dir, utils.JitDefaultPermission); err != nil {
		return err
	}

	var compressed_data bytes.Buffer
	zw := zlib.NewWriter(&compressed_data)

	if _, err := zw.Write(data); err != nil {
		return err
	}

	zw.Close()

	// temp_file, err := os.CreateTemp(object_dir, db.generateTmpObjectName(oid_hex))
	temp_file, err := os.CreateTemp(object_dir, db.generateTmpObjectName(oid))

	if err != nil {
		return err
	}

	if _, err = temp_file.Write(compressed_data.Bytes()); err != nil {
		return err
	}

	temp_file.Close()

	return os.Rename(temp_file.Name(), object_path)
}

func (db *Database) generateTmpObjectName(oid []byte) string {

	return fmt.Sprintf("tmp_obj_%x", oid[0:3])
}

func (db *Database) GetChanges() int {
	return db.count
}
func (db *Database) HashObject(object database.Object) ([]byte, error) {
	content, err := db.serializeObject(object)
	if err != nil {
		return nil, err
	}
	return db.hashContent(content), nil
}

func (db *Database) serializeObject(object database.Object) ([]byte, error) {
	data := []byte(fmt.Sprintf("%s %d\x00", object.Type(), len(object.ToString())))
	data = append(data, object.ToString()...)

	var encoded_data bytes.Buffer
	if err := binary.Write(&encoded_data, binary.BigEndian, data); err != nil {
		return nil, err
	}

	return encoded_data.Bytes(), nil
}

func (db *Database) hashContent(content []byte) []byte {
	objectId := sha1.Sum(content)
	return objectId[:]
}

func (db *Database) objectPath(oid []byte) string {
	oid_hex := fmt.Sprintf("%x", oid)
	return filepath.Join(db.path, oid_hex[:2], oid_hex[2:])
}

/*
	So far we were writing data to our database
	Now we want to load the data from the database
	For a commit, we take the tree oid from it and load it from the db
	then from this tree, we loop on its entries
		if the current entry is a tree, we recurse on it
		if the current entry is a blob, we return its oid, mode, ....

	So now we need to implement a function called load and it takes a [] byte oid
	and returns an object for us

*/

// func (db *Database) load(oid []byte) database.Object {

// }

// for now, we will make sure that we could get the type and the size of the current object
func (db *Database) readObject(oid []byte) {
	// objectPath := db.objectPath(oid)
	// compressedData, err := os.ReadFile(objectPath)
	// if err != nil {
	// 	return "", -1, err
	// }

	// reader, err := zlib.NewReader(bytes.NewReader(compressedData))
	// if err != nil {
	// 	return "", -1, err
	// }

	// defer reader.Close()

	// data, err := io.ReadAll(reader)
	// if err != nil {
	// 	return "", -1, err
	// }

	// spaceIdx := bytes.IndexByte(data, ' ')
	// nullIdx := bytes.IndexByte(data, 0)

	// if spaceIdx == -1 || nullIdx == -1 {
	// 	return "", -1, fmt.Errorf("invalid object format: no space or null terminator found")
	// }

	// objectType := string(data[:spaceIdx])
	// sizeStr := string(data[spaceIdx+1 : nullIdx])
	// objectSize, err := strconv.Atoi(sizeStr)

	// if err != nil {
	// 	return "", -1, fmt.Errorf("invalid object size: %v", err)
	// }

	// return objectType, objectSize, nil
}
