package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"
)

const (
	DEFAULT_CONFIG_FILE = "config.toml"
)

var (
	logger             = ligneous.NewLogger()
	PORT        int    = 5555
	config      Config = Config{}
	CONFIG_FILE string = DEFAULT_CONFIG_FILE
)

func init() {
	config.Fetch(CONFIG_FILE)
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/v1/layers", LayersHandler).Methods("GET")
	router.HandleFunc("/v1/layer/{layer_name}", LayerHandler).Methods("GET")
	router.HandleFunc("/v1/tile/{layer_name}/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.mvt", VectorTileHandler).Methods("GET")

	router.Use(LoggingMiddleWare, SetHeadersMiddleWare)

	port := fmt.Sprintf("%v", PORT)
	logger.Info(fmt.Sprintf("Magic happens on port %v...", port))
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}

}
