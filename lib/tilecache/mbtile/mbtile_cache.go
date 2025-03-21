package mbtiles

import (
	"fmt"
	"sync"

	"mvt-server/lib/mbtiles"
)

const DEFAULT_TILE_CACHE_DIRECTORY string = "cache"

var TILE_CACHE_DIRECTORY = DEFAULT_TILE_CACHE_DIRECTORY

func New(directory string) (*MBTileCache, error) {
    return &MBTileCache{directory: directory, databases: make(map[string]*mbtiles.Database)}, nil
}

type MBTileCache struct {
	directory string
	databases      map[string]*mbtiles.Database
	sync.RWMutex
}

func (self *MBTileCache) getDirectory() string {
	if "" != self.directory {
		return TILE_CACHE_DIRECTORY
	}
	return self.directory
}

func (self *MBTileCache) Init() {
	self.Lock()
	defer self.Unlock()

	if nil == self.databases {
		self.databases = make(map[string]*mbtiles.Database)
	}
}

func (self *MBTileCache) newDb(layerName string) (*mbtiles.Database, error) {
	self.Lock()
	defer self.Unlock()

	filePath := fmt.Sprintf("%v/%v.mbtiles", self.getDirectory(), layerName)
	db, err := mbtiles.NewDatabase(filePath)
	if nil != err {
		return db, err
	}

	self.databases[filePath] = db
	return db, nil
}

func (self *MBTileCache) getDb(layerName string) (*mbtiles.Database, error) {
	self.RLock()

	db, ok := self.databases[layerName]
	if !ok {
		self.RUnlock()
		return self.newDb(layerName)
	}

	self.RUnlock()
	return db, nil
}

func (self *MBTileCache) GetTile(layerName string, x, y, z uint32) ([]uint8, error) {
	db, err := self.getDb(layerName)
	if nil != err {
		return []uint8{}, err
	}
	return db.GetTile(z, x, y)
}

func (self *MBTileCache) SetTile(layerName string, x, y, z uint32, tileData []uint8) error {
	db, err := self.getDb(layerName)
	if nil != err {
		return err
	}
	return db.SetTile(z, x, y, tileData)
}

func (self *MBTileCache) SetMetadata(layerName string, metadata [][]string) error {
	// bounds := layer.Extent.Geometry().Bound()
	// center := bounds.Center()
	// metadata := [][]string{
	// 	[]string{"name", layer.LayerName},
	// 	[]string{"format", "pbf"},
	// 	[]string{"type", "overlay"},
	// 	[]string{"description", layer.Description},
	// 	[]string{"attribution", layer.Attribution},
	// 	[]string{"version", "1.0"},
	// 	[]string{"bounds", fmt.Sprintf("%v,%v,%v,%v", bounds.Left(), bounds.Bottom(), bounds.Right(), bounds.Top())},
	// 	[]string{"center", fmt.Sprintf("%v,%v,0", center.X(), center.Y())},
	// 	[]string{"minzoom", "0"},
	// 	[]string{"maxzoom", "20"},
	// 	[]string{"json", fmt.Sprintf(`{"vector_layers":[{"id":"%v","description":"%v","minzoom":0,"maxzoom":20}]}`, layer.LayerName, layer.Description)},
	// }

	db, err := self.getDb(layerName)
	if nil != err {
		return err
	}
	return db.InsertMetadata(metadata)
}
