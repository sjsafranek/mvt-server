package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"

	"mvt-server/lib/tilecache"
)

const (
	DEFAULT_ACTION      string = "web"
	DEFAULT_CONFIG_FILE string = "config.toml"
	DEFAULT_PORT        int    = 5555
	PROJECT             string = "mvt-server"
	VERSION             string = "0.01.11"
)

var (
	logger      = ligneous.AddLogger("app", "debug", "logs")
	tileCache   tilecache.TileCache
	config      Config = Config{}
	CONFIG_FILE string = DEFAULT_CONFIG_FILE
	PORT        int    = DEFAULT_PORT
	DEBUG       bool   = false
	ACTION      string = DEFAULT_ACTION
)

func init() {
	var print_version bool

	flag.StringVar(&ACTION, "action", DEFAULT_ACTION, "Action")
	flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "Config file")
	flag.IntVar(&PORT, "p", DEFAULT_PORT, "Server port")
	flag.BoolVar(&DEBUG, "debug", false, "debug mode")
	flag.StringVar(&DATABASE_HOST, "h", DEFAULT_DATABASE_HOST, "database host")
	flag.StringVar(&DATABASE_DATABASE, "n", DEFAULT_DATABASE_DATABASE, "database name")
	flag.StringVar(&DATABASE_PASSWORD, "pw", DEFAULT_DATABASE_PASSWORD, "database password")
	flag.StringVar(&DATABASE_USERNAME, "un", DEFAULT_DATABASE_USERNAME, "database username")
	flag.Int64Var(&DATABASE_PORT, "dbp", DEFAULT_DATABASE_PORT, "Database port")
	flag.StringVar(&tilecache.TILE_CACHE_TYPE, "cacheType", tilecache.DEFAULT_TILE_CACHE_TYPE, "tile cache type")

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

	layers, err := NewLayerCollection(config.Database.ConnectionString(), false)
	if nil != err {
		panic(err)
	}
	LAYERS = layers

	switch ACTION {

	case "cook":

		cache, err := tilecache.NewTileCache(config.Cache)
		if nil != err {
			panic(err)
		}
		tileCache = cache

		args := flag.Args()
		layer := args[0]
		filter := args[1]
		beginZoom, _ := strconv.ParseUint(args[2], 10, 64)
		endZoom, _ := strconv.ParseUint(args[3], 10, 64)

		// allow user to specify their own bounding box
		if 8 == len(args) {
			minlat, _ := strconv.ParseFloat(args[4], 64)
			maxlat, _ := strconv.ParseFloat(args[5], 64)
			minlng, _ := strconv.ParseFloat(args[6], 64)
			maxlng, _ := strconv.ParseFloat(args[7], 64)

			CookTilesInSelectedBounds(layer, filter, int(beginZoom), int(endZoom), minlat, maxlat, minlng, maxlng)
		} else {
			// use bounding box of layer (default)
			CookTilesInLayerBounds(layer, filter, int(beginZoom), int(endZoom))
		}

	case "web":

		logger.Debugf("%v:%v", PROJECT, VERSION)
		hostname, err := os.Hostname()
		logger.Debug("Hostname: ", hostname)
		logger.Debug("GOOS: ", runtime.GOOS)
		logger.Debug("CPUS: ", runtime.NumCPU())
		logger.Debug("PID: ", os.Getpid())
		logger.Debug("Go Version: ", runtime.Version())
		logger.Debug("Go Arch: ", runtime.GOARCH)
		logger.Debug("Go Compiler: ", runtime.Compiler)
		logger.Debug("NumGoroutine: ", runtime.NumGoroutine())

		cache, err := tilecache.NewTileCache(config.Cache)
		if nil != err {
			panic(err)
		}
		tileCache = cache

		// Why wasn't this handled?
		// _ = LAYERS.Init()
		go func() {
			err = LAYERS.Init()
			if nil != err {
				panic(err)
			}
		}()

		go commandListener()

		router := mux.NewRouter()
		// https://wiki.osgeo.org/wiki/Tile_Map_Service_Specification
		router.HandleFunc("/web", ViewMapHandler).Methods("GET", "OPTIONS")
		router.HandleFunc("/api/v1/layers", ApiV1LayersHandler).Methods("GET", "OPTIONS")
		router.HandleFunc("/api/v1/layer/{layer_name}", ApiV1LayerHandler).Methods("GET", "OPTIONS")
		router.HandleFunc("/api/v1/layer/{layer_name}/wfs", ApiV1WebFeatureServiceHandler).Methods("POST", "OPTIONS")
		router.HandleFunc("/api/v1/layer/{layer_name}/tile/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.mvt", ApiV1TileMapServiceGetVectorTile).Methods("GET", "OPTIONS")

		router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

		router.Use(LoggingMiddleWare(logger), SetHeadersMiddleWare, CORSMiddleWare)

		port := fmt.Sprintf("%v", PORT)

		logger.Info(fmt.Sprintf("Magic happens on port %v...", port))
		err = http.ListenAndServe(":"+port, router)
		if err != nil {
			panic(err)
		}

	case "upload":

		_ = LAYERS.Init()

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
