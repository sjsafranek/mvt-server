package main

import (
	"sync"
	"time"

	"mvt-server/lib/tileutils"
)

func tileWorker(layer *Layer, filterSQL string, queue chan tileutils.Tile, wg *sync.WaitGroup) {
	for tile := range queue {
		start := time.Now()
		tileData, err := layer.FetchTile(tile.X, tile.Y, tile.Z, filterSQL)
		if nil != err {
			logger.Error(err)
			continue
		}
		logger.Infof("Got tile - %v bytes - %v", len(tileData), time.Since(start))
	}
	wg.Done()
}

func CookTilesInSelectedBounds(layerName, filterJSON string, beginZoom, endZoom int, minlat, maxlat, minlng, maxlng float64) {
	err := LAYERS.AddLayer(layerName)
	if nil != err {
		panic(err)
	}

	layer, err := LAYERS.GetLayer(layerName)
	if nil != err {
		panic(err)
	}

	wg := sync.WaitGroup{}
	queue := make(chan tileutils.Tile, 4)
	filterSQL, err := formatFilter(filterJSON, layer)
	if nil != err {
		panic(err)
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go tileWorker(layer, filterSQL, queue, &wg)
	}

	for zoom := beginZoom; zoom < endZoom; zoom++ {
		xyzs := tileutils.GetTilesFromBounds(minlat, maxlat, minlng, maxlng, zoom)
		for i := range xyzs {
			queue <- xyzs[i]
		}
	}
	close(queue)

	wg.Wait()
}

func CookTilesInLayerBounds(layerName, filterJSON string, beginZoom, endZoom int) {
	err := LAYERS.AddLayer(layerName)
	if nil != err {
		panic(err)
	}

	layer, err := LAYERS.GetLayer(layerName)
	if nil != err {
		panic(err)
	}

	bbox := layer.Extent.Geometry().Bound()

	minlat := bbox.Min[1]
	maxlat := bbox.Max[1]
	minlng := bbox.Min[0]
	maxlng := bbox.Max[0]

	CookTilesInSelectedBounds(layerName, filterJSON, beginZoom, endZoom, minlat, maxlat, minlng, maxlng)

	// wg := sync.WaitGroup{}
	// queue := make(chan tileutils.Tile, 4)
	// filterSQL, err := formatFilter(filterJSON, layer)
	// if nil != err {
	// 	panic(err)
	// }
	//
	// for i := 0; i < 4; i++ {
	// 	wg.Add(1)
	// 	go tileWorker(layer, filterSQL, queue, &wg)
	// }
	//
	// for zoom := beginZoom; zoom < endZoom; zoom++ {
	// 	xyzs := tileutils.GetTilesFromBounds(minlat, maxlat, minlng, maxlng, zoom)
	// 	for i := range xyzs {
	// 		queue <- xyzs[i]
	// 	}
	// }
	// close(queue)
	//
	// wg.Wait()
}
