package main

import (
	"sync"
	"time"
)

func tileWorker(layerName string, queue chan xyz, wg *sync.WaitGroup) {
	for xyz := range queue {
		start := time.Now()
		tile := NewTile(layerName, xyz.x, xyz.y, xyz.z)
		tileData, err := tile.Fetch()
		if nil != err {
			logger.Error(err)
			continue
		}
		logger.Infof("Got tile - %v bytes - %v", len(tileData), time.Since(start))
	}
	wg.Done()
}

func CookTiles(layerName string) {
	bbox := LAYERS[layerName].Extent.Geometry().Bound()

	minlat := bbox.Min[1]
	maxlat := bbox.Max[1]
	minlng := bbox.Min[0]
	maxlng := bbox.Max[0]

	wg := sync.WaitGroup{}
	queue := make(chan xyz, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go tileWorker(layerName, queue, &wg)
	}

	for zoom := 0; zoom < 14; zoom++ {
		xyzs := GetTileNamesFromMapView(minlat, maxlat, minlng, maxlng, zoom)
		for i := range xyzs {
			queue <- xyzs[i]
		}
	}
	close(queue)

	wg.Wait()
}
