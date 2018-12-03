package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"errors"

	"github.com/paulmach/orb/maptile"
)

var (
	cacheDir = "cache"
	emptyTile = []uint8{}
)

func NewTile(layer_name string, x, y, z uint32) Tile {
	tile := maptile.New(x, y, maptile.Zoom(z))
	return Tile{Layer: layer_name, X: x, Y: y, Z: z, MapTile: tile}
}

type Tile struct {
	Layer   string
	X       uint32
	Y       uint32
	Z       uint32
	MapTile maptile.Tile
	MVT     []uint8
}

func (self *Tile) getTileName() string {
	return fmt.Sprintf("%v/%v/%v/%v.mvt", self.Layer, self.Z, self.X, self.Y)
}

func (self *Tile) isEmpty() bool {
	return !self.MapTile.Bound().Intersects(LAYERS[self.Layer])
}

func (self *Tile) Fetch() ([]uint8, error) {

	if _, ok := LAYERS[self.Layer]; !ok {
		return emptyTile, errors.New("Layer not found")
	}

	tileName := self.getTileName()

	if self.isEmpty() {
		logger.Infof("Empty tile %v - 0 bytes", tileName)
		return emptyTile, nil
	}

	tileData, err := self.fetchTileFromCache()
	if nil == err {
		logger.Infof("Got tile %v from cache - %v bytes", tileName, len(tileData))
		return tileData, err
	}

	tileData, err = fetchTileFromDatabase(self.Layer, self.X, self.Y, self.Z)
	if nil != err {
		logger.Error(err)
		return tileData, err
	}

	logger.Infof("Got tile %v from database - %v bytes", tileName, len(tileData))
	go self.cacheTile(tileData)

	return tileData, err
}

func (self *Tile) fetchTileFromCache() ([]uint8, error) {
	tileName := self.getTileName()
	tileFile := fmt.Sprintf("%v/%v", cacheDir, tileName)
	tileData, err := ioutil.ReadFile(tileFile)
	return tileData, err
}

func (self *Tile) cacheTile(tileData []uint8) error {
	tileName := self.getTileName()
	tileDirPath := fmt.Sprintf("%v/%v/%v/%v", cacheDir, self.Layer, self.Z, self.X)
	tileFile := fmt.Sprintf("%v/%v", cacheDir, tileName)

	err := os.MkdirAll(tileDirPath, os.ModePerm)
	if nil != err {
		panic(err)
	}

	f, err := os.Create(tileFile)
	if nil != err {
		return err
	}
	defer f.Close()

	_, err = f.Write(tileData)
	return err
}
