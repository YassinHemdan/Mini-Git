package internals

import (
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
	Store(Object) error
}

type Database struct {
	path string
}

func (db *Database) New(path string) error {
	db.path = path
	fmt.Println("New database created at", db.path)

	return nil
}

func (db *Database) Store(object Object) error {
	// prepare the data that we want to encode and hash
	data := []byte(fmt.Sprintf("%s %d\x00", object.Type(), len(object.ToString())))
	data = append(data, object.ToString()...)

	// encode the data
	var encoded_data bytes.Buffer
	if err := binary.Write(&encoded_data, binary.BigEndian, data); err != nil {
		return err
	}

	// hash the data using sha1sum
	objectId := sha1.Sum(encoded_data.Bytes())
	object.SetOid(objectId[:])

	return db.writeObject(object.GetOid(), encoded_data.Bytes())
}

func (db *Database) writeObject(oid, data []byte) error {
	oid_hex := fmt.Sprintf("%x", oid)
	object_dir := strings.Join([]string{db.path, oid_hex[:2]}, string(os.PathSeparator))     //.git/objects/xx
	object_path := strings.Join([]string{object_dir, oid_hex[2:]}, string(os.PathSeparator)) // .git/objects/xx/ooooooo

	// if a file exists already, don't write it again
	if _, err := os.Stat(object_path); err == nil {
		return nil
	}

	if err := os.MkdirAll(object_dir, JitDefaultPermission); err != nil {
		return err
	}

	// compress the data

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

	// we write our compressed data to the temp file
	if _, err = temp_file.Write(compressed_data.Bytes()); err != nil {
		return err
	}

	temp_file.Close()

	// move the temp file to the obj file

	fmt.Println(object_path) // just for debugging purposes

	return os.Rename(temp_file.Name(), object_path)
}

func (db *Database) generateTmpObjectName(hex_oid string) string {

	return fmt.Sprintf("tmp_obj_%x", hex_oid[0:3])
}
