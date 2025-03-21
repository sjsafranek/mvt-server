package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/sjsafranek/goutils/hashers"

	"mvt-server/lib/geodatabase"
	"mvt-server/lib/tilecache"
)

var (
	EMPTY_TILE = []uint8{}
)

type LayerColumn struct {
	ColumnId string `json:"column_id"`
	Type     string `json:"type"`
}

type Layer struct {
	// TODO: ompit empty?? extent --
	Extent        geojson.Polygon          `json:"extent,omitempty"`
	Features      int                      `json:"features"`
	SRID          int64                    `json:"srid"`
	LayerId       string                   `json:"layer_id"`
	CreatedAt     string                   `json:"created_at"`
	IsDeleted     bool                     `json:"is_deleted"`
	LayerName     string                   `json:"layer_name"`
	UpdatedAt     string                   `json:"updated_at"`
	Description   string                   `json:"description"`
	Attribution   string                   `json:"attribution"`
	IsUpdatable   bool                     `json:"is_updatable"` // allow live updates to layers
	Columns       []LayerColumn            `json:"columns"`
	GeometryTypes []string                 `json:"geometry_types"`
	geodb         *geodatabase.GeoDatabase `json:"-"` // remove from json response
	tileCache     tilecache.TileCache      `json:"-"` // remove from json response
	// Properties    []string                 `json:"properties"`
}

func (self *Layer) FetchTile(x, y, z uint32, filter string) ([]uint8, error) {
	filterHsh := hashers.MD5HashString(filter)
	tileName := fmt.Sprintf("layer:%v filter:%v z:%v x:%v y:%v", self.LayerName, filterHsh, z, x, y)

	// check for empty tile
	tile := maptile.New(x, y, maptile.Zoom(z))
	if !tile.Bound().Intersects(self.Extent.Geometry().Bound()) {
		logger.Debugf("Empty tile %v", tileName)
		return EMPTY_TILE, nil
	}

	// check tile cache
	lyrName := fmt.Sprintf("%v_%v", self.LayerName, filterHsh)
	tileData, err := tileCache.GetTile(lyrName, x, y, z)
	if nil == err {
		logger.Infof("Got tile %v from cache", tileName)
		return tileData, err
	}

	// fetch from database
	tileData, err = self.geodb.FetchTile(self.LayerName, x, y, z, self.SRID, filter)
	if nil != err {
		logger.Error(err)
		return tileData, err
	}

	logger.Infof("Got tile %v from database", tileName)

	// save to tile cache
	go tileCache.SetTile(lyrName, x, y, z, tileData)

	return tileData, err
}

func (self *Layer) FetchTileWithContext(ctx context.Context, x, y, z uint32, filter string) ([]uint8, error) {
	filterHsh := hashers.MD5HashString(filter)
	tileName := fmt.Sprintf("layer:%v filter:%v z:%v x:%v y:%v", self.LayerName, filterHsh, z, x, y)

	// check for empty tile
	//  - if layer is 'is_updatable' the bounds are changing
	tile := maptile.New(x, y, maptile.Zoom(z))
	if !self.IsUpdatable && !tile.Bound().Intersects(self.Extent.Geometry().Bound()) {
		logger.Debugf("Empty tile %v", tileName)
		return EMPTY_TILE, nil
	}

	// check tile cache
	lyrName := fmt.Sprintf("%v_%v", self.LayerName, filterHsh)
	tileData, err := self.tileCache.GetTile(lyrName, x, y, z)
	if nil == err {
		logger.Infof("Got tile %v from cache", tileName)
		return tileData, err
	}

	// fetch from database
	tileData, err = self.geodb.FetchTileWithContext(ctx, self.LayerName, x, y, z, self.SRID, filter)
	if nil != err {
		return tileData, err
	}

	logger.Infof("Got tile %v from database", tileName)

	// save to tile cache
	if !self.IsUpdatable {
		go self.tileCache.SetTile(lyrName, x, y, z, tileData)
	}

	return tileData, err
}

func (self *Layer) Delete() error {
	logger.Warnf("Delete layer %v", self.LayerName)
	return self.geodb.DeleteLayer(self.LayerName)
}

func (self *Layer) QueryRow(query string, results ...interface{}) error {
	return self.geodb.QueryRow(query, &results)
}

func (self *Layer) QueryRowJSON(query string) (string, error) {
	return self.geodb.QueryRowJSON(query)
}

func (self *Layer) MarshalToJSON() ([]byte, error) {
	return json.Marshal(self)
}
