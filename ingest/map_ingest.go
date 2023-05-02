package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/packetflinger/q2mapinfo/common"
)

var (
	dbfile = flag.String("db", "mapdata.sqlite", "The sqlite3 database file")
	force  = flag.Bool("force", false, "Remove existing maps and insert the new one")
)

// insert a map into the system.
func ingestMap(m string, db *sql.DB) (*common.BSPMap, error) {
	bsp := &common.BSPMap{}
	bsp.Database = db
	bsp.Filepath = m
	t := strings.Split(m, "/")
	name := t[len(t)-1]           // get last field
	bsp.Name = name[:len(name)-4] // remove ".bsp" from end

	bsp.OpenMap()
	id := bsp.GetMapFromHash()
	tx, err := bsp.StartTransaction()
	bsp.Transaction = tx
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if id > 0 {
		if *force {
			e := bsp.Delete(id)
			if e != nil {
				tx.Rollback()
				return bsp, e
			}
		} else {
			log.Printf("skipping %s [%s], exists (force with --force flag)\n", bsp.Name, bsp.Hash)
			tx.Rollback()
			return bsp, errors.New("existing map")
		}
	}

	bsp.ParseTextures()
	bsp.ParseEntities()

	res, err := bsp.Database.Exec("INSERT INTO map (name, hash, dateadded) VALUES (?, ?, ?)", bsp.Name, bsp.Hash, common.Now())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	bsp.ID, err = res.LastInsertId()
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return nil, err
	}

	e := bsp.InsertTextures()
	if e != nil {
		tx.Rollback()
		return nil, err
	}

	e = bsp.InsertEntities()
	if e != nil {
		tx.Rollback()
		return nil, err
	}

	return bsp, nil
}

func main() {
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		fmt.Printf("Usage: %s <map.bsp> [map.bsp...]\n", os.Args[0])
		os.Exit(0)
	}

	db := common.OpenDatabase(*dbfile)
	for _, m := range files {
		_, e := ingestMap(m, db)
		if e != nil {
			log.Println(e)
		}
	}
	db.Close()
}
