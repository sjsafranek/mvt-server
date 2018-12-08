package main

import (
	"encoding/json"
	"errors"

	// "github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

var (
	LAYERS Layers
)

type Layers map[string]LayerMetadata

func (self *Layers) LayerExists(layer_name string) bool {
	_, ok := LAYERS[layer_name]
	return ok
}

func (self *Layers) GetLayer(layer_name string) (LayerMetadata, error) {
	layer, ok := LAYERS[layer_name]
	if !ok {
		return layer, errors.New("Layer not found")
	}
	return layer, nil
}

type LayerMetadata struct {
	Extent      geojson.Polygon `json:"extent"`
	Features    int             `json:"features"`
	SRID        int64           `json:"srid"`
	LayerId     string          `json:"layer_id"`
	CreatedAt   string          `json:"created_at"`
	IsDeleted   bool            `json:"is_deleted"`
	LayerName   string          `json:"layer_name"`
	UpdatedAt   string          `json:"updated_at"`
	Description string          `json:"description"`
}

// loadLayerMetadata loads layer bounds from database
func loadLayerMetadata() {
	logger.Debug("Loading layer metadata")

	LAYERS = Layers{}

	var layers []map[string]interface{}
	res, err := fetchLayersFromDatabase()
	if nil != err {
		panic(err)
	}

	err = json.Unmarshal([]byte(res), &layers)
	if nil != err {
		panic(err)
	}

	for i := range layers {
		layer_name := layers[i]["layer_name"].(string)
		logger.Debugf("Fetch %v metadata", layer_name)

		var lyrMetadata LayerMetadata
		res, err = fetchLayerFromDatabase(layer_name)
		if nil != err {
			logger.Error(err)
			// logger.Warnf("Marking layer %v as deleted", layer_name)
			// deleteLayerFromDatabase(layer_name)
			// panic(err)
			continue
		}

		err = json.Unmarshal([]byte(res), &lyrMetadata)
		if nil != err {
			panic(err)
		}

		LAYERS[layer_name] = lyrMetadata
	}

}
