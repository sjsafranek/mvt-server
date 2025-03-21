package mbtiles

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

/*
	Database
	Based off of -->
	https://github.com/twpayne/go-mbtiles
*/

const MBTILE_TABLES string = `
		BEGIN TRANSACTION;

		CREATE TABLE IF NOT EXISTS metadata (
			name TEXT NOT NULL UNIQUE,
			value TEXT
		);

		CREATE TABLE IF NOT EXISTS tiles (
			zoom_level INTEGER NOT NULL,
			tile_column INTEGER NOT NULL,
			tile_row INTEGER NOT NULL,
			tile_data BLOB NOT NULL
		);

		CREATE UNIQUE INDEX IF NOT EXISTS tile_index
			ON tiles (zoom_level, tile_column, tile_row);

		COMMIT;
	`

func NewDatabase(db_file string) (*Database, error) {
	db, err := sql.Open("sqlite3", db_file)
	if nil != err {
		return &Database{}, err
	}
	mbt := Database{db: db}
	return &mbt, nil
}

type Database struct {
	insertQueue    chan func() error
	db             *sql.DB
	tileInsertStmt *sql.Stmt
	tileSelectStmt *sql.Stmt
}

func (self *Database) Open(filePath string) error {
	if nil == self.db {
		db, err := sql.Open("sqlite3", filePath)
		if err != nil {
			return err
		}
		self.db = db
	}
	return nil
}

func (self *Database) Close() error {
	var err error
	if nil != self.db {
		err = self.db.Close()
		if nil != err {

			if nil != self.tileInsertStmt {
				err = self.tileInsertStmt.Close()
			}

			if nil != self.tileSelectStmt {
				err = self.tileSelectStmt.Close()
			}

		}
	}
	return err
}

func (self *Database) createTables() error {
	// create tables
	_, err := self.db.Exec(MBTILE_TABLES)
	return err
}

func (self *Database) InsertMetadata(metadata [][]string) error {
	// insert metadata
	tx, err := self.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO metadata(name, value) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, row := range metadata {
		_, err = stmt.Exec(row[0], row[1])
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func (self *Database) init() error {
	if nil != self.db {

		// open database connection
		err := self.Open(fmt.Sprintf("./db.Database"))
		if err != nil {
			return err
		}

		// create tables
		if err := self.createTables(); nil != err {
			return err
		}

		// cannot do concurrent writes with sqlite3
		self.insertQueue = make(chan func() error, 8)
		go func() {

			if nil == self.tileInsertStmt {
				var err error
				self.tileInsertStmt, err = self.db.Prepare("INSERT OR REPLACE INTO tiles (zoom_level, tile_column, tile_row, tile_data) VALUES (?, ?, ?, ?);")
				if err != nil {
					panic(err)
				}
			}

			for clbk := range self.insertQueue {
				err := clbk()
				if err != nil {
					panic(err)
				}
			}
		}()
		//.end

	}
	return nil
}

// InsertTile inserts a tile at (z, x, y).
func (self *Database) InsertTile(z, x, y uint32, tileData []byte) error {
	if err := self.init(); nil != err {
		return err
	}

	// cannot do concurrent writes with sqlite3
	var err error
	self.insertQueue <- func() error {
		_, err2 := self.tileInsertStmt.Exec(z, x, y, tileData)
		err = err2
		return err2
	}
	return err
}

func (self *Database) SetTile(z, x, y uint32, tileData []byte) error {
	return self.InsertTile(z, x, y, tileData)
}

// SelectTile returns the tile at (z, x, y).
func (self *Database) SelectTile(z, x, y uint32) ([]byte, error) {
	if err := self.init(); nil != err {
		return nil, err
	}

	if self.tileSelectStmt == nil {
		var err error
		self.tileSelectStmt, err = self.db.Prepare("SELECT tile_data FROM tiles WHERE zoom_level = ? AND tile_column = ? AND tile_row = ?;")
		if err != nil {
			return nil, err
		}
	}

	var tileData []byte
	// https://github.com/twpayne/go-Database/blob/master/reader.go
	// err := self.tileSelectStmt.QueryRow(z, x, 1<<uint(z)-y-1).Scan(&tileData)
	err := self.tileSelectStmt.QueryRow(z, x, y).Scan(&tileData)
	return tileData, err
}

func (self *Database) GetTile(z, x, y uint32) ([]byte, error) {
	return self.SelectTile(z, x, y)
}
