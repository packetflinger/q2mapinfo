package main

import (
	"fmt"
	"os"
)

const (
	EntLump = 0 // the location in the header
)

/**
 * Find the entity lump in the BSP.
 * Return the location and length
 */
func LocateEntityLump(header []byte) (int, int) {
	var offsets [19]int
	var lengths [19]int

	pos := 8
	for i := 0; i < 18; i++ {
		offsets[i] = int(ReadLong(header, pos)) - HeaderLen
		pos = pos + 4
		lengths[i] = int(ReadLong(header, pos))
		pos = pos + 4
	}

	return offsets[EntLump] + HeaderLen, lengths[EntLump]
}

/**
 * Get a slice of the just the texture lump from the map file
 */
func GetEntityLump(f *os.File, offset int, length int) []byte {
	_, err := f.Seek(int64(offset), 0)
	check(err)

	lump := make([]byte, length)
	read, err := f.Read(lump)
	check(err)

	if read != length {
		panic("reading entity lump: hit EOF")
	}

	return lump
}

func ParseEntities() {
	bspname := os.Args[1]
	bsp, err := os.Open(bspname)
	check(err)

	header := make([]byte, HeaderLen)
	_, err = bsp.Read(header)
	check(err)

	VerifyHeader(header)

	offset, length := LocateEntityLump(header)
	ents := GetEntityLump(bsp, offset, length)
	fmt.Println(string(ents))
}
