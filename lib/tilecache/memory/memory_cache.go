package memory

import (
	"errors"
	"fmt"
	"sync"
)

// https://github.com/go-spatial/tegola/blob/master/cache/memory/memory.go

func New() (*MemoryTileCache, error) {
    return &MemoryTileCache{tiles: make(map[string][]byte)}, nil
}

type MemoryTileCache struct {
	tiles map[string][]byte
	sync.RWMutex
}

func (self *MemoryTileCache) Init() {
    self.Lock()
	defer self.Unlock()
    if nil == self.tiles {
    	self.tiles = make(map[string][]byte)
    }
}

func (self *MemoryTileCache) GetTile(layerName string, x, y, z uint32) ([]uint8, error) {
    self.RLock()
	defer self.RUnlock()

	key := fmt.Sprintf("%v-%v-%v-%v", layerName, z, x, y)
	tileData, ok := self.tiles[key]
	if !ok {
		return tileData, errors.New("Not found")
	}

	return tileData, nil
}

func (self *MemoryTileCache) SetTile(layerName string, x, y, z uint32, tileData []uint8) error {
    self.Lock()
	defer self.Unlock()

	key := fmt.Sprintf("%v-%v-%v-%v", layerName, z, x, y)
	self.tiles[key] = tileData

	return nil
}

func (self *MemoryTileCache) SetMetadata(layerName string, metadata [][]string) error {
	return nil
}
