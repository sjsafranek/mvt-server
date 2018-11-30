package main

import (
	"fmt"
    "strconv"
	"net/http"

	"github.com/gorilla/mux"
    "github.com/sjsafranek/ligneous"
	"github.com/paulmach/orb/encoding/mvt"
)

var logger = ligneous.NewLogger()
var PORT = 5555


func VectorTileHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
	layer_name := vars["layer_name"]
	z,_ := strconv.Atoi(vars["z"])
	x,_ := strconv.Atoi(vars["x"])
    y,_ := strconv.Atoi(vars["y"])

    b, err := fetchTile(layer_name, x, y, z)

	layers, err := mvt.Unmarshal(b)
	if nil != err {
		panic(err)
	}

    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Max-Age", "86400")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Content-Type", "application/octet-stream")

    logger.Infof("%v features in tile", len(layers[0].Features))

    w.Write(b)
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



/*

http://localhost:5555/v1/tile/tl_2017_us_puma/6/14/24.mvt



L.vectorGrid.protobuf("http://localhost:5555/v1/tile/tl_2017_us_puma/{z}/{x}/{y}.mvt", {
    //vectorTileLayerStyles: { ... },
}).addTo(
    app.maps[0].LMap
);



*/
