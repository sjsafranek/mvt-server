package nocache

import (
	"errors"
)

func New() (*NoCache, error) {
	return &NoCache{}, nil
}

type NoCache struct {}

func (self *NoCache) GetTile(layerName string, x, y, z uint32) ([]uint8, error) {
	return []uint8{}, errors.New("Not found")
}

func (self *NoCache) SetTile(layerName string, x, y, z uint32, tileData []uint8) error {
	return nil
}

func (self *NoCache) SetMetadata(layerName string, metadata [][]string) error {
	return nil
}
