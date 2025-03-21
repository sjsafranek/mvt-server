package disk

import (
	"fmt"
	"io/ioutil"
	"os"
)

const DEFAULT_TILE_CACHE_DIRECTORY string = "cache"

var TILE_CACHE_DIRECTORY = DEFAULT_TILE_CACHE_DIRECTORY

func New(directory string) (*DiskTileCache, error) {
    return &DiskTileCache{directory: directory}, nil
}

type DiskTileCache struct {
	directory string
}

func (self *DiskTileCache) getDirectory() string {
	if "" != self.directory {
		return TILE_CACHE_DIRECTORY
	}
	return self.directory
}

func (self *DiskTileCache) GetTile(layerName string, x, y, z uint32) ([]uint8, error) {
	tileName := fmt.Sprintf("%v/%v/%v/%v.mvt", layerName, z, x, y)
	tileFile := fmt.Sprintf("%v/%v", self.getDirectory(), tileName)
	tileData, err := ioutil.ReadFile(tileFile)
	return tileData, err
}

func (self *DiskTileCache) SetTile(layerName string, x, y, z uint32, tileData []uint8) error {
	tileDirPath := fmt.Sprintf("%v/%v/%v/%v", self.getDirectory(), layerName, z, x)
	err := os.MkdirAll(tileDirPath, os.ModePerm)
	if nil != err {
		panic(err)
	}

	tileName := fmt.Sprintf("%v/%v/%v/%v.mvt", layerName, z, x, y)
	tileFile := fmt.Sprintf("%v/%v", self.getDirectory(), tileName)
	f, err := os.Create(tileFile)
	if nil != err {
		return err
	}
	defer f.Close()

	_, err = f.Write(tileData)
	return err
}

func (self *DiskTileCache) SetMetadata(layerName string, metadata [][]string) error {
	fmt.Println(layerName, metadata)
	return nil
}
