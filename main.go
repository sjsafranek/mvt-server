package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"
)

const (
	DEFAULT_ACTION      string = "web"
	DEFAULT_CONFIG_FILE string = "config.toml"
	DEFAULT_PORT        int    = 5555
	PROJECT             string = "mvt-server"
	VERSION             string = "0.0.1"
)

var (
	logger             = ligneous.NewLogger()
	config      Config = Config{}
	CONFIG_FILE string = DEFAULT_CONFIG_FILE
	PORT        int    = DEFAULT_PORT
	ACTION      string = DEFAULT_ACTION
)

func init() {
	var print_version bool

	flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "Server port")
	flag.IntVar(&PORT, "p", DEFAULT_PORT, "Server port")
	flag.BoolVar(&print_version, "V", false, "Print version and exit")
	flag.Parse()

	if print_version {
		fmt.Println(PROJECT, VERSION)
		os.Exit(0)
	}

	err := config.Fetch(CONFIG_FILE)
	if nil != err {
		panic(err)
	}

	if 2 <= len(os.Args) {
		ACTION = os.Args[1]
		if "web" != ACTION && "upload" != ACTION {
			panic(errors.New("Please specifiy action [web(default), upload]"))
		}

	}
}

func main() {

	switch ACTION {
	case "web":

		loadLayerMetadata()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/layers", LayersHandler).Methods("GET")
		router.HandleFunc("/api/v1/layer/{layer_name}", LayerHandler).Methods("GET")
		router.HandleFunc("/api/v1/layer/{layer_name}/tile/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.mvt", VectorTileHandler).Methods("GET")

		router.Use(LoggingMiddleWare, SetHeadersMiddleWare)

		port := fmt.Sprintf("%v", PORT)
		logger.Info(fmt.Sprintf("Magic happens on port %v...", port))
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			panic(err)
		}

	case "upload":
		// TODO: check for errors...
		shapefile := os.Args[2]
		description := os.Args[3]
		srid, _ := strconv.ParseInt(os.Args[4], 10, 64)
		res, err := UploadShapefile(shapefile, description, srid)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)

	default:
		panic(errors.New("Please specifiy action [web(default), upload]"))
	}

}
