package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// TODO
//  - fix database creation

// https://mapserver.org/mapcache/caches.html#mapcache-cache-sqlite

const SQLITE3_TABLES string = `
		BEGIN TRANSACTION;

        CREATE TABLE IF NOT EXISTS tiles(
           tileset TEXT,
           grid TEXT,
           x INTEGER,
           y INTEGER,
           z INTEGER,
           data BLOB,
           dim TEXT,
           ctime DATETIME,
           PRIMARY KEY(tileset,grid,x,y,z,dim)
        );

		COMMIT;
	`

const DEFAULT_TILE_CACHE_DIRECTORY string = "cache"

var TILE_CACHE_DIRECTORY = DEFAULT_TILE_CACHE_DIRECTORY

func New(directory string) (*SQLiteTileCache, error) {
    return &SQLiteTileCache{directory: directory}, nil
}

type SQLiteTileCache struct {
	directory      string
	insertQueue    chan func() error
	db             *sql.DB
	tileInsertStmt *sql.Stmt
	tileSelectStmt *sql.Stmt
}

func (self *SQLiteTileCache) getDirectory() string {
	if "" != self.directory {
		return TILE_CACHE_DIRECTORY
	}
	return self.directory
}

func (self *SQLiteTileCache) createTables() error {
	// create tables
	_, err := self.db.Exec(SQLITE3_TABLES)
	return err
}

func (self *SQLiteTileCache) open(db_file string) error {
	filePath := fmt.Sprintf("./%v/%v", self.getDirectory(), db_file)
	db, err := sql.Open("sqlite3", filePath)
	self.db = db
	if nil == err {
		err = self.createTables()

		// cannot do concurrent writes with sqlite3
		self.insertQueue = make(chan func() error, 8)
		go func() {

			if nil == self.tileInsertStmt {
				var err error
				self.tileInsertStmt, err = self.db.Prepare("INSERT OR REPLACE INTO tiles (z, x, y, data, tileset) VALUES (?, ?, ?, ?, ?);")
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
	return err
}

func (self *SQLiteTileCache) Close() error {
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

func (self *SQLiteTileCache) GetTile(layerName string, x, y, z uint32) ([]uint8, error) {
	if nil == self.db {
		err := self.open("cache.sqlite3")
		if nil != err {
			return []uint8{}, err
		}
	}

	if self.tileSelectStmt == nil {
		var err error
		self.tileSelectStmt, err = self.db.Prepare("SELECT data FROM tiles WHERE z = ? AND x = ? AND y = ? AND tileset = ?;")
		if err != nil {
			return nil, err
		}
	}

	var tileData []byte
	// https://github.com/twpayne/go-Database/blob/master/reader.go
	// err := self.tileSelectStmt.QueryRow(z, x, 1<<uint(z)-y-1).Scan(&tileData)
	err := self.tileSelectStmt.QueryRow(z, x, y, layerName).Scan(&tileData)
	return tileData, err

}

func (self *SQLiteTileCache) SetTile(layerName string, x, y, z uint32, tileData []uint8) error {
	// cannot do concurrent writes with sqlite3
	var err error
	self.insertQueue <- func() error {
		_, err2 := self.tileInsertStmt.Exec(z, x, y, tileData, layerName)
		err = err2
		return err2
	}
	return err
}

func (self *SQLiteTileCache) SetMetadata(layerName string, metadata [][]string) error {
	return nil
}
