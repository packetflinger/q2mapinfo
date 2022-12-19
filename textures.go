package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

const (
	Magic       = (('P' << 24) + ('S' << 16) + ('B' << 8) + 'I')
	HeaderLen   = 160 // magic + version + lump metadata
	TextureLump = 5   // the location in the header
	TextureLen  = 76  // 40 bytes of origins and angles + 36 for textname
)

/**
 * Just simple error checking
 */
func check(e error) {
	if e != nil {
		panic(e)
	}
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

/**
 * Make sure the first 4 bytes match the magic number
 */
func VerifyHeader(header []byte) {
	if ReadLong(header, 0) != Magic {
		panic("Invalid BPS file")
	}
}

/**
 * Find the offset and the length of the texture lump
 * in the BSP file
 */
func LocateTextureLump(header []byte) (int, int) {
	var offsets [19]int
	var lengths [19]int

	pos := 8
	for i := 0; i < 18; i++ {
		offsets[i] = int(ReadLong(header, pos)) - HeaderLen
		pos = pos + 4
		lengths[i] = int(ReadLong(header, pos))
		pos = pos + 4
	}

	return offsets[TextureLump] + HeaderLen, lengths[TextureLump]
}

/**
 * Get a slice of the just the texture lump from the map file
 */
func GetTextureLump(f *os.File, offset int, length int) []byte {
	_, err := f.Seek(int64(offset), 0)
	check(err)

	lump := make([]byte, length)
	read, err := f.Read(lump)
	check(err)

	if read != length {
		panic("reading texture lump: hit EOF")
	}

	return lump
}

/**
 * Loop through all the textures in the lump building a
 * slice of just the texture names
 */
func GetTextures(lump []byte) []string {
	size := len(lump) / TextureLen
	var textures []string
	pos := 0
	for i := 0; i < size; i++ {
		pos += 40
		texture := lump[pos : pos+32]
		pos += 32 + 4
		textures = append(textures, string(texture))
	}

	return textures
}

/**
 * Remove any duplipcates
 */
func Deduplicate(in []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range in {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

/**
 *
 */
func ParseTextures() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <q2map.bsp> [q2map.bsp...]\n", os.Args[0])
		return
	}

	textures := []string{}

	for _, bspname := range os.Args[1:] {
		bsp, err := os.Open(bspname)
		check(err)

		header := make([]byte, HeaderLen)
		_, err = bsp.Read(header)
		check(err)

		VerifyHeader(header)

		offset, length := LocateTextureLump(header)
		texturelump := GetTextureLump(bsp, offset, length)
		textures = append(textures, GetTextures(texturelump)...)

		bsp.Close()
	}

	dedupedtextures := Deduplicate(textures)
	sort.Strings(dedupedtextures)

	for _, t := range dedupedtextures {
		fmt.Println(t)
	}
}
