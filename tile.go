package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

var cacheDir = "cache"

type Tile struct {
	Layer string
	X int
	Y int
	Z int
	MVT []uint8
}

func (self *Tile) getTileName() string {
	return fmt.Sprintf("%v/%v/%v/%v.mvt", self.Layer, self.X, self.Y, self.Z)
}

func (self *Tile) Fetch() ([]uint8, error) {
	tileName := self.getTileName()

	tileData, err := self.fetchTileFromCache()
	if nil == err {
		logger.Debugf("Got tile %v from cache", tileName)
		return tileData, err
	}

	tileData, err = fetchTileFromDatabase(self.Layer, self.X, self.Y, self.Z)
	if nil == err {
		go self.cacheTile(tileData)
	}

	logger.Debugf("Got tile %v from database", tileName)
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
	tileDirPath := fmt.Sprintf("%v/%v/%v/%v", cacheDir, self.Layer, self.X, self.Y)
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
