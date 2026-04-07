package internals

// lets import the Object

import (
	database "JIT/internals/database"
	"JIT/internals/utils"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type IDatabase interface {
	New(string) error
	Store(database.Object) error // we need to import it
}

type Database struct {
	count int
	path  string
}

func (db *Database) New(path string) error {
	db.count = 0
	db.path = path
	fmt.Println("New database created at", db.path)

	return nil
}

func (db *Database) Store(object database.Object) error {
	// objToStr := object
	data := []byte(fmt.Sprintf("%s %d\x00", object.Type(), len(object.ToString())))
	data = append(data, object.ToString()...)

	var encoded_data bytes.Buffer
	if err := binary.Write(&encoded_data, binary.BigEndian, data); err != nil {
		return err
	}

	objectId := sha1.Sum(encoded_data.Bytes())
	object.SetOid(objectId[:])

	return db.writeObject(object.GetOid(), encoded_data.Bytes())
}

func (db *Database) writeObject(oid, data []byte) error {
	oid_hex := fmt.Sprintf("%x", oid)
	object_dir := strings.Join([]string{db.path, oid_hex[:2]}, string(os.PathSeparator))
	object_path := strings.Join([]string{object_dir, oid_hex[2:]}, string(os.PathSeparator))

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

	temp_file, err := os.CreateTemp(object_dir, db.generateTmpObjectName(oid_hex))

	if err != nil {
		return err
	}

	if _, err = temp_file.Write(compressed_data.Bytes()); err != nil {
		return err
	}

	temp_file.Close()

	return os.Rename(temp_file.Name(), object_path)
}

func (db *Database) generateTmpObjectName(hex_oid string) string {

	return fmt.Sprintf("tmp_obj_%x", hex_oid[0:3])
}

func (db *Database) GetChanges() int {
	return db.count
}
