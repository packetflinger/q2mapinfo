package common

import (
	"os"
	"sort"
	"strings"
)

// Find the offset and the length of the texture lump
// in the BSP file
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

// Get a slice of the just the texture lump from the map file
func GetTextureLump(f *os.File, offset int, length int) []byte {
	_, err := f.Seek(int64(offset), 0)
	Check(err)

	lump := make([]byte, length)
	read, err := f.Read(lump)
	Check(err)

	if read != length {
		panic("reading texture lump: hit EOF")
	}

	return lump
}

// Loop through all the textures in the lump building a
// slice of just the texture names
func GetTextures(lump []byte) []string {
	size := len(lump) / TextureLen
	var textures []string
	pos := 0
	for i := 0; i < size; i++ {
		pos += 40
		texture := lump[pos : pos+32]
		pos += 32 + 4
		textures = append(textures, strings.Trim(string(texture), "\x00"))
	}

	return textures
}

// Remove any duplipcates
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
func (bspmap *BSPMap) ParseTextures() {
	textures := []string{}

	bsp := bspmap.FilePointer

	offset, length := LocateTextureLump(bspmap.Header)
	texturelump := GetTextureLump(bsp, offset, length)
	textures = append(textures, GetTextures(texturelump)...)

	dedupedtextures := Deduplicate(textures)
	sort.Strings(dedupedtextures)

	bspmap.Textures = dedupedtextures
}
