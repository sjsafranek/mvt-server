package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
		z, _ := strconv.Atoi(vars["z"])
		x, _ := strconv.Atoi(vars["x"])
		y, _ := strconv.Atoi(vars["y"])

		tile := Tile{Layer: layer_name, X: x, Y: y, Z: z}
		tileData, err := tile.Fetch()
		if nil != err {
			jsonHttpResponse(w, 500, err.Error())
			return queue
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(tileData)
		return queue
	}():
		logger.Debugf("request completed %v", time.Since(start))

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

	layer = strings.Replace(layer, `\"`, `"`, -1)
	layer = strings.Replace(layer, `"{`, `{`, -1)
	layer = strings.Replace(layer, `}"`, `}`, -1)
	// layer = layer[:len(layer)-2]

	jsonHttpResponse(w, 200, fmt.Sprintf(`{"layer": %v}`, layer))
}
