package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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

	flag.StringVar(&ACTION, "action", DEFAULT_ACTION, "Action")
	flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "Config file")
	flag.IntVar(&PORT, "p", DEFAULT_PORT, "Server port")
	flag.BoolVar(&print_version, "V", false, "Print version and exit")
	flag.Parse()

	if print_version {
		fmt.Println(PROJECT, VERSION)
		os.Exit(0)
	}

	err := config.Fetch(CONFIG_FILE)
	if nil != err {
		logger.Warn(err)
		logger.Info("Using default config settings")
		err = config.UseDefaults()
		if nil != err {
			panic(err)
		}
	}

	signal_queue := make(chan os.Signal)
	signal.Notify(signal_queue, syscall.SIGTERM)
	signal.Notify(signal_queue, syscall.SIGINT)
	go func() {
		sig := <-signal_queue
		logger.Warnf("caught sig: %+v", sig)
		logger.Warn("Gracefully shutting down...")
		logger.Warn("Shutting down...")
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()

}

func main() {

	logger.Infof("Using database connection: %v", config.Database.ConnectionString())

	switch ACTION {

	case "ls":
		loadLayerMetadata()
		for layer := range LAYERS {
			fmt.Println(layer)
		}

	case "cook":
		loadLayerMetadata()
		args := flag.Args()
		layer := args[0]
		beginZoom, _ := strconv.ParseUint(args[1], 10, 64)
		endZoom, _ := strconv.ParseUint(args[2], 10, 64)
		CookTiles(layer, int(beginZoom), int(endZoom))

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
		if 4 != len(flag.Args()) {
			fmt.Println(flag.Args())
			fmt.Println("Incorrect usage: <shapefile> <tablename> <description> <srid>")
			return
		}

		args := flag.Args()
		shapefile := args[0]
		tablename := args[1]
		description := args[2]
		srid, _ := strconv.ParseInt(args[3], 10, 64)
		res, err := UploadShapefile(shapefile, tablename, description, srid)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)

	default:
		fmt.Printf(`Invalid option.

Usage:
	%v [action] [options]

Actions:
	web (default)		Start http server
	upload			Upload shapefile

Options:

\n`, os.Args[0])
	}

}
