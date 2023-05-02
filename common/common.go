package common

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	Magic       = (('P' << 24) + ('S' << 16) + ('B' << 8) + 'I')
	HeaderLen   = 160 // magic + version + lump metadata
	TextureLump = 5   // the location in the header
	TextureLen  = 76  // 40 bytes of origins and angles + 36 for textname
	EntLump     = 0   // the location in the header
)

type BSPMap struct {
	Name        string
	ID          int64 // db key
	Filepath    string
	FilePointer *os.File
	Hash        string
	Header      []byte
	Textures    []string
	Entities    []Entity
	EntCounts   map[string]int
	Database    *sql.DB
	Transaction *sql.Tx
}

// Read 4 bytes as a Long
func ReadLong(input []byte, start int) int32 {
	var tmp struct {
		Value int32
	}

	r := bytes.NewReader(input[start:])
	if err := binary.Read(r, binary.LittleEndian, &tmp); err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	return tmp.Value
}

// Just simple error checking
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// opens the file and saves the pointer for later
func (bspmap *BSPMap) OpenMap() {
	fp, err := os.Open(bspmap.Filepath)
	Check(err)

	bspmap.FilePointer = fp
	VerifyHeader(bspmap)
	CalcMD5Sum(bspmap)
}

func (bspmap *BSPMap) CloseMap() {
	bspmap.FilePointer.Close()
}

// Make sure the first 4 bytes match the magic number
func VerifyHeader(bspmap *BSPMap) {
	bspmap.Header = make([]byte, HeaderLen)
	_, err := bspmap.FilePointer.Read(bspmap.Header)
	Check(err)

	if ReadLong(bspmap.Header, 0) != Magic {
		panic("Invalid BPS file")
	}
}

// get a unix timestamp for use in the database
func Now() int64 {
	return time.Now().Unix()
}

// calculate a hash of the map file
func CalcMD5Sum(bspmap *BSPMap) error {
	hash := md5.New()

	if _, err := io.Copy(hash, bspmap.FilePointer); err != nil {
		return err
	}

	bytes := hash.Sum(nil)[:16]
	bspmap.Hash = hex.EncodeToString(bytes)
	return nil
}

// Remove all data related to a map
func (bspmap *BSPMap) Delete(id int) error {
	//tx := bspmap.Transaction
	ctx := context.Background()
	//if tx == nil {
	tx, err := bspmap.Database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	//	bspmap.Transaction = tx
	//}
	fmt.Println(tx)
	_, err = tx.ExecContext(ctx, "DELETE FROM map WHERE name = ?", bspmap.Name)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM texture WHERE map = ?", bspmap.ID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM entity WHERE map = ?", bspmap.ID)
	if err != nil {
		return err
	}

	return nil
}
