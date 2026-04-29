package internals

import (
	// "JIT/internals/database"
	database "JIT/internals/database"
	"JIT/internals/utils"
	scanner "JIT/utils"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type IDatabase interface {
	New(string) error
	Store(database.Object) error
}

type ParseFunc func(scanner *scanner.SmartScanner) database.Object

var TYPES = map[string]ParseFunc{
	"blob":   database.ParseBlob,
	"commit": database.ParseCommit,
	"tree":   database.ParseTree,
}

type Database struct {
	count         int
	path          string
	cachedObjects map[string]database.Object
}

func NewDatabase(pathname string) (*Database, error) {
	return &Database{
		count:         0,
		path:          pathname,
		cachedObjects: make(map[string]database.Object),
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

func (db *Database) Load(oid []byte) (database.Object, error) {
	object, ok := db.cachedObjects[string(oid)]
	if ok {
		return object, nil
	}

	object, err := db.readObject(oid)
	if err != nil {
		return nil, err
	}

	db.cachedObjects[string(oid)] = object
	return object, nil
}

func (db *Database) readObject(oid []byte) (database.Object, error) {
	objectPath := db.objectPath(oid)
	// fmt.Println(objectPath)
	compressedData, err := os.ReadFile(objectPath)
	if err != nil {
		return nil, fmt.Errorf("Error: Could not read compressedData - %v", err)
	}

	reader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("Error: Could not create a new reader for compressed data - %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Error: Could not ReadAll data - %v", err)
	}
	scanner := scanner.NewObjectScanner(bytes.NewReader(data))
	scanner.SplitByDelim(' ', true)
	scanner.Scan()
	objectType := scanner.Text() // type (commit, tree, blob)

	scanner.SplitByDelim('\x00', true)
	scanner.Scan()
	scanner.Text() // size

	object := TYPES[objectType](scanner)
	object.SetOid(oid)
	return object, nil
}
