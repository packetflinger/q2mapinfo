package main

import (
	"flag"
)

// flags
var (
	Verbose      = flag.Bool("v", false, "Show extra info")
	Databasefile = flag.String("database", "./mapdata.sqlite", "Database file to use")
)

/*
func mainold() {
	flag.Parse()

	if *Verbose {
		log.Printf("Database: %s\n", *Databasefile)
	}

	for _, file := range flag.Args() {
		bsp := &common.BSPMap{}
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
			common.Check(err)

			bsp.ID, err = res.LastInsertId()
			common.Check(err)
		}

		bsp.ParseTextures()
		bsp.ParseEntities()

		bsp.InsertTextures()
		bsp.InsertEntities()

		bsp.CloseMap()
		bsp.CloseDatabase()
	}
}
*/
