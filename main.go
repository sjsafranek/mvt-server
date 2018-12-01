package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"
	// "github.com/paulmach/orb/encoding/mvt"
)

var logger = ligneous.NewLogger()
var PORT = 5555

// func VectorTileHandler(w http.ResponseWriter, r *http.Request) {
//
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 	    w.Header().Set("Access-Control-Max-Age", "86400")
// 	    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
// 	    w.Header().Set("Access-Control-Allow-Credentials", "true")
//
// 	    vars := mux.Vars(r)
// 		layer_name := vars["layer_name"]
// 		z,_ := strconv.Atoi(vars["z"])
// 		x,_ := strconv.Atoi(vars["x"])
// 	    y,_ := strconv.Atoi(vars["y"])
//
// 	    tileData, err := fetchTile(ctx, layer_name, x, y, z)
// 		if nil != err {
// 			w.Header().Set("Content-Type", "application/json")
// 			w.WriteHeader(500)
// 			payload := fmt.Sprintf(`{"status":"error","error": "%v"}`, err.Error())
// 			fmt.Fprintln(w, payload)
// 			return
// 		}
//
// 		// layers, err := mvt.Unmarshal(tileData)
// 		// if nil != err {
// 		// 	panic(err)
// 		// }
//
// 		w.Header().Set("Content-Type", "application/octet-stream")
// 	    w.Write(tileData)
//
// }

func VectorTileHandler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	// logger.Info("request initiated")

	ctx := r.Context()

	select {

	case <-func() <-chan bool {
		queue := make(chan bool, 1)
		queue <- true

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		vars := mux.Vars(r)
		layer_name := vars["layer_name"]
		z, _ := strconv.Atoi(vars["z"])
		x, _ := strconv.Atoi(vars["x"])
		y, _ := strconv.Atoi(vars["y"])

		tile := Tile{Layer: layer_name, X: x, Y: y, Z: z}
		tileData, err := tile.Fetch()
		if nil != err {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			payload := fmt.Sprintf(`{"status":"error","error": "%v"}`, err.Error())
			fmt.Fprintln(w, payload)
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

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/v1/tile/{layer_name}/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.mvt", VectorTileHandler).Methods("GET")

	port := fmt.Sprintf("%v", PORT)
	logger.Info(fmt.Sprintf("Magic happens on port %v...", port))
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}

}
