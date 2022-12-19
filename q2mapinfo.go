package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
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
	Verbose = flag.Bool("v", false, "Show extra info")
)

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

/**
 * Make sure the first 4 bytes match the magic number
 */
func VerifyHeader(header []byte) {
	if ReadLong(header, 0) != Magic {
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

func main() {
	flag.Parse()

	for _, file := range flag.Args() {
		ParseTextures(file)
		ParseEntities(file)
	}
}
