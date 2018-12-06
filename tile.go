package main

import (
	// "errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/paulmach/orb/maptile"
	"github.com/sjsafranek/goutils/hashers"
)

var (
	cacheDir  = "cache"
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
	Filter  string
}

func (self *Tile) getTileName() string {
	return fmt.Sprintf("%v/%v/%v/%v_%v.mvt", self.Layer, self.Z, self.X, self.Y, hashers.MD5HashString(self.Filter))
}

func (self *Tile) isEmpty() bool {
	layer, err := LAYERS.GetLayer(self.Layer)
	if nil != err {
		return true
	}
	return !self.MapTile.Bound().Intersects(layer.Extent.Geometry().Bound())
}

func (self *Tile) Fetch() ([]uint8, error) {

	tileName := self.getTileName()

	if self.isEmpty() {
		logger.Infof("Empty tile %v", tileName)
		return emptyTile, nil
	}

	tileData, err := self.fetchTileFromCache()
	if nil == err {
		logger.Infof("Got tile %v from cache", tileName)
		return tileData, err
	}

	tileData, err = fetchTileFromDatabase(self.Layer, self.X, self.Y, self.Z, self.Filter)
	if nil != err {
		logger.Error(err)
		return tileData, err
	}

	logger.Infof("Got tile %v from database", tileName)
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
