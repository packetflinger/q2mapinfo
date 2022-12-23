package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const (
	Magic       = (('P' << 24) + ('S' << 16) + ('B' << 8) + 'I')
	HeaderLen   = 160 // magic + version + lump metadata
	TextureLump = 5   // the location in the header
	TextureLen  = 76  // 40 bytes of origins and angles + 36 for textname
	EntLump     = 0   // the location in the header
)

// flags
var (
	Verbose      = flag.Bool("v", false, "Show extra info")
	Databasefile = flag.String("database", "./mapdata.sqlite", "Database file to use")
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
}

/**
 * Read 4 bytes as a Long
 */
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

// opens the file and saves the pointer for later
func (bspmap *BSPMap) OpenMap() {
	fp, err := os.Open(bspmap.Filepath)
	check(err)

	bspmap.FilePointer = fp
	bspmap.VerifyHeader()
	bspmap.CalcMD5Sum()
}

func (bspmap *BSPMap) CloseMap() {
	bspmap.FilePointer.Close()
}

/**
 * Make sure the first 4 bytes match the magic number
 */
func (bspmap *BSPMap) VerifyHeader() {
	bspmap.Header = make([]byte, HeaderLen)
	_, err := bspmap.FilePointer.Read(bspmap.Header)
	check(err)

	if ReadLong(bspmap.Header, 0) != Magic {
		panic("Invalid BPS file")
	}
}

/**
 * Just simple error checking
 */
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// get a unix timestamp for use in the database
func Now() int64 {
	return time.Now().Unix()
}

// calculate a hash of the map file
func (bspmap *BSPMap) CalcMD5Sum() error {
	hash := md5.New()

	if _, err := io.Copy(hash, bspmap.FilePointer); err != nil {
		panic(err)
	}

	bytes := hash.Sum(nil)[:16]
	bspmap.Hash = hex.EncodeToString(bytes)
	return nil
}

func main() {
	flag.Parse()

	if *Verbose {
		log.Printf("Database: %s\n", *Databasefile)
	}

	for _, file := range flag.Args() {
		bsp := &BSPMap{}
		bsp.Filepath = file

		fields := strings.Split(file, "/")

		name := fields[len(fields)-1] // get last field
		bsp.Name = name[:len(name)-4] // remove ".bsp" from end
		bsp.OpenMap()
		bsp.OpenDatabase()

		id := bsp.GetMapFromHash()

		if id > 0 {
			log.Printf("Skipping %s [%s], already in database\n", bsp.Name, bsp.Hash)
			continue // already in the db, skip it
		} else {
			res, err := bsp.Database.Exec("INSERT INTO map (name, hash, dateadded) VALUES (?, ?, ?)", bsp.Name, bsp.Hash, Now())
			check(err)

			bsp.ID, err = res.LastInsertId()
			check(err)
		}

		bsp.ParseTextures()
		bsp.ParseEntities()

		bsp.InsertTextures()
		bsp.InsertEntities()

		bsp.CloseMap()
		bsp.CloseDatabase()
	}
}
