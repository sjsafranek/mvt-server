package main

import (
	"fmt"
	"net/http"
	"strconv"
	// "strings"
	"time"

	"github.com/gorilla/mux"
	// "github.com/paulmach/orb/encoding/mvt"
)

func jsonHttpResponse(w http.ResponseWriter, status_code int, result string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status_code)
	var payload string
	if 300 <= status_code {
		payload = fmt.Sprintf(`{"status":"error","error": "%v"}`, result)
		logger.Error(payload)
	} else {
		payload = fmt.Sprintf(`{"status":"ok","data": %v}`, result)
		logger.Debug(payload)
	}
	fmt.Fprintln(w, payload)
}

func VectorTileHandler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	ctx := r.Context()

	select {

	case <-func() <-chan bool {
		queue := make(chan bool, 1)
		queue <- true

		vars := mux.Vars(r)
		layer_name := vars["layer_name"]
		z, _ := strconv.ParseUint(vars["z"], 10, 64)
		x, _ := strconv.ParseUint(vars["x"], 10, 64)
		y, _ := strconv.ParseUint(vars["y"], 10, 64)

		tile := NewTile(layer_name, uint32(x), uint32(y), uint32(z))
		tileData, err := tile.Fetch()
		if nil != err {
			jsonHttpResponse(w, 500, err.Error())
			return queue
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(tileData)
		return queue
	}():
		logger.Tracef("request completed %v", time.Since(start))

	case <-ctx.Done():
		logger.Warnf("request cancelled %v", time.Since(start))
	}

}

func LayersHandler(w http.ResponseWriter, r *http.Request) {
	layers, err := fetchLayersFromDatabase()
	if nil != err {
		jsonHttpResponse(w, 500, err.Error())
		return
	}

	jsonHttpResponse(w, 200, fmt.Sprintf(`{"layers": %v}`, layers))
}

func LayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	layer_name := vars["layer_name"]
	layer, err := fetchLayerFromDatabase(layer_name)
	if nil != err {
		jsonHttpResponse(w, 500, err.Error())
		return
	}

	jsonHttpResponse(w, 200, fmt.Sprintf(`{"layer": %v}`, layer))
}
