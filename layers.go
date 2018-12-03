package main

import (
	"encoding/json"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

var (
	LAYERS map[string]orb.Bound
)

type LayerMetadata struct {
	Extent   geojson.Polygon `json:"extent"`
	Features int             `json:"features"`
}

// loadLayerMetadata loads layer bounds from database
func loadLayerMetadata() {
	logger.Debug("Loading layer metadata")

	LAYERS = make(map[string]orb.Bound)

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

		var layer LayerMetadata
		res, err = fetchLayerFromDatabase(layer_name)
		if nil != err {
			panic(err)
		}

		err = json.Unmarshal([]byte(res), &layer)
		if nil != err {
			panic(err)
		}

		LAYERS[layer_name] = layer.Extent.Geometry().Bound()
	}

}

func layerExists(layer_name string) bool {
	_, ok := LAYERS[layer_name]
	return ok
}
