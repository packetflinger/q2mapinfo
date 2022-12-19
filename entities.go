package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Possible fields we care about
type Entity struct {
	Classname  string
	Message    string
	Sky        string
	Origin     string
	Angle      string
	Spawnflags string
	Noise      string
	Light      string
	Targetname string
	Model      string
	Height     string
	Wait       string
	Speed      string
	Accel      string
}

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

func BreakupEntityLump(lump []byte) []Entity {
	ents := []Entity{}
	current := Entity{}
	inside := false
	lines := strings.Split(string(lump), "\n")
	for _, line := range lines {
		if !inside && line == "{" {
			inside = true
			current = Entity{}
			continue
		}

		if inside && line == "}" {
			inside = false
			ents = append(ents, current)
			continue
		}

		if inside {
			re, err := regexp.Compile(" ")
			check(err)
			keyval := re.Split(line, 2)
			key := keyval[0][1 : len(keyval[0])-1]
			val := keyval[1][1 : len(keyval[1])-1]
			fmt.Printf("%s = %s\n", key, val)

			switch key {
			case "classname":
				current.Classname = val
			case "origin":
				current.Origin = val
			}
		}
	}
	return ents
}

func ParseEntities(file string) {
	bsp, err := os.Open(file)
	check(err)

	header := make([]byte, HeaderLen)
	_, err = bsp.Read(header)
	check(err)

	VerifyHeader(header)

	offset, length := LocateEntityLump(header)
	lump := GetEntityLump(bsp, offset, length)
	ents := BreakupEntityLump(lump)
	fmt.Println(ents)
}
