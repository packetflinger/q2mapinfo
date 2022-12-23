package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func (bsp *BSPMap) OpenDatabase() {
	db, err := sql.Open("sqlite3", *Databasefile)
	if err != nil {
		panic(err)
	}

	bsp.Database = db
}

func (bsp *BSPMap) CloseDatabase() {
	bsp.Database.Close()
}

func (bsp *BSPMap) DBAddTexture(t string) {
	_, err := bsp.Database.Exec("INSERT INTO texture (map, texture) VALUES (?,?)", bsp.ID, t)
	check(err)
}

func (bsp *BSPMap) DBAddEntity(c string, q int) {
	_, err := bsp.Database.Exec("INSERT INTO entity (map, classname, quantity) VALUES (?,?,?)", bsp.ID, c, q)
	check(err)
}

func (bsp *BSPMap) GetMapFromHash() int {
	res, err := bsp.Database.Query("SELECT id FROM map WHERE hash = ? LIMIT 1", bsp.Hash)
	check(err)
	defer res.Close()

	id := 0

	for res.Next() {
		err := res.Scan(&id)
		if err == sql.ErrNoRows {
			return 0
		}
	}
	return id
}

func (bsp *BSPMap) InsertTextures() {
	// first delete any textures we already have for this map
	_, err := bsp.Database.Exec("DELETE FROM texture WHERE map = ?", bsp.ID)
	check(err)

	for _, t := range bsp.Textures {
		bsp.DBAddTexture(t)
	}
}

func (bsp *BSPMap) InsertEntities() {
	_, err := bsp.Database.Exec("DELETE FROM entity WHERE map = ?", bsp.ID)
	check(err)

	for k, v := range bsp.EntCounts {
		bsp.DBAddEntity(k, v)
	}
}

func (bsp *BSPMap) GetTexture_test() string {
	row, err := bsp.Database.Query("SELECT texture FROM texture WHERE id = 1045 LIMIT 1")
	check(err)

	defer row.Close()
	t := ""
	for row.Next() {
		err := row.Scan(&t)
		if err == sql.ErrNoRows {
			return ""
		}
	}
	return t
}
