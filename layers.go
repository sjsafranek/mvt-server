package main

import (
	"encoding/json"
	// "errors"
	"fmt"
	"strings"
	"sync"

	"mvt-server/lib/geodatabase"
	"mvt-server/lib/tilecache"
)

var (
	LAYERS *LayerCollection
)

type LayerCollection struct {
	geodb  *geodatabase.GeoDatabase `json:"-"`
	layers map[string]*Layer        `json:"layers"`
	guard  sync.RWMutex             `json:"-"`
}

func (self *LayerCollection) LayerExists(layerName string) bool {
	self.guard.RLock()
	_, ok := self.layers[layerName]
	self.guard.RUnlock()
	return ok
}

func (self *LayerCollection) AddLayer(layerName string) error {
	// fix case sensitive
	layerName = strings.ToLower(layerName)

	var layer Layer
	res, err := self.geodb.FetchLayer(layerName)
	if nil != err {
		return err
	}

	// // is this need?
	// if "" == res {
	// 	return errors.New("Layer not found: " + layerName)
	// }
	// //.end

	if DEBUG {
		fmt.Println(layerName, res)
	}

	// HACK:
	// github.com/paulmach/orb/geojson.(*Polygon).UnmarshalJSON
	// causes a panic...
	// runtime error: invalid memory address or nil pointer dereference
	// [signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x7612f3]
	if strings.Contains(res, `"extent": null,`) {
		return nil
	}
	//.end

	err = json.Unmarshal([]byte(res), &layer)
	if nil != err {
		return err
	}

	if nil == layer.Extent {
		return nil
	}

	self.addLayer(layerName, layer)
	return nil
}

func (self *LayerCollection) addLayer(layerName string, layer Layer) {
	self.guard.Lock()
	if nil == self.layers {
		self.layers = make(map[string]*Layer)
	}
	logger.Infof("Add layer %v", layerName)
	layer.geodb = self.geodb
	if !layer.IsUpdatable {
		layer.tileCache = tileCache
	} else {
		nocache, _ := tilecache.NewTileCache(tilecache.Config{Type: "none"})
		layer.tileCache = nocache
	}
	self.layers[layerName] = &layer
	self.guard.Unlock()

	//
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
	// tileCache.SetMetadata(layer, metadata)
	//
}

func (self *LayerCollection) GetLayer(layerName string) (*Layer, error) {
	self.guard.RLock()
	defer self.guard.RUnlock()
	layer, ok := self.layers[layerName]
	if !ok {
		return layer, fmt.Errorf("Layer %v not found", layerName)
	}
	return layer, nil
}

func (self *LayerCollection) DeleteLayer(layerName string) error {
	layer, err := self.GetLayer(layerName)
	if nil != err {
		return err
	}

	self.guard.Lock()
	defer self.guard.Unlock()

	layer.Delete()
	delete(self.layers, layerName)

	return nil
}

func (self *LayerCollection) Init() error {
	var layers []map[string]interface{}

	logger.Debug("Loading layers")
	res, err := self.geodb.FetchLayers()
	if nil != err {
		return err
	}

	if "" == res {
		logger.Warn("No layers found")
		return nil
	}

	err = json.Unmarshal([]byte(res), &layers)
	if nil != err {
		return err
	}

	for i := range layers {
		layerName := layers[i]["layer_name"].(string)
		logger.Debugf("Fetch layer %v", layerName)
		// if debug slow query
		if DEBUG {
			self.AddLayer(layerName)
		} else {
			go self.AddLayer(layerName)
		}
	}

	return nil
}

func (self *LayerCollection) FetchLayersFromDatabase() (string, error) {
	return self.geodb.FetchLayers()
}

func (self *LayerCollection) FetchLayerFromDatabase(layerName string) (string, error) {
	return self.geodb.FetchLayer(layerName)
}

// loadLayerMetadata loads layer bounds from database
func NewLayerCollection(dsName string, loadAllLayersFromDatabase bool) (*LayerCollection, error) {
	geodb, err := geodatabase.NewGeoDatabase(dsName)
	if nil != err {
		return &LayerCollection{}, err
	}
	geodb.Debug = DEBUG

	collection := LayerCollection{geodb: geodb}

	if loadAllLayersFromDatabase {
		err = collection.Init()
	}

	return &collection, err
}

func (self *LayerCollection) MarshalToJSON() ([]byte, error) {
	layers := []*Layer{}
	self.guard.RLock()
	for layer_name := range self.layers {
		layers = append(layers, self.layers[layer_name])
	}
	self.guard.RUnlock()
	return json.Marshal(layers)
}
