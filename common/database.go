package common

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDatabase(dbfile string) *sql.DB {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		panic(err)
	}

	//bsp.Database = db
	return db
}

func (bsp *BSPMap) CloseDatabase() {
	bsp.Database.Close()
}

func (bsp *BSPMap) DBAddTexture(t string) error {
	_, err := bsp.Database.Exec("INSERT INTO texture (map, texture) VALUES (?,?)", bsp.ID, t)
	if err != nil {
		return err
	}
	return nil
}

func (bsp *BSPMap) DBAddEntity(c string, q int) error {
	_, err := bsp.Database.Exec("INSERT INTO entity (map, classname, quantity) VALUES (?,?,?)", bsp.ID, c, q)
	if err != nil {
		return err
	}
	return nil
}

func (bsp *BSPMap) GetMapFromHash() int {
	res, err := bsp.Database.Query("SELECT id FROM map WHERE hash = ? LIMIT 1", bsp.Hash)
	Check(err)
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

func (bsp *BSPMap) StartTransaction() (*sql.Tx, error) {
	t, err := bsp.Database.Begin()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return t, nil
}

func (bsp *BSPMap) InsertTextures() error {
	for _, t := range bsp.Textures {
		e := bsp.DBAddTexture(t)
		if e != nil {
			return e
		}
	}
	return nil
}

func (bsp *BSPMap) InsertEntities() error {
	for k, v := range bsp.EntCounts {
		e := bsp.DBAddEntity(k, v)
		if e != nil {
			return e
		}
	}
	return nil
}

func (bsp *BSPMap) GetTexture_test() string {
	row, err := bsp.Database.Query("SELECT texture FROM texture WHERE id = 1045 LIMIT 1")
	Check(err)

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
