package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/ligneous"
	"mvt-server/lib/tilecache"
)

type Config struct {
	Layers map[string]string `json:"layers"`
	Port   int               `json:"port"`
}

const (
	DEFAULT_MAX_CONCURRENT_JOBS int = 4
)

var (
	logger            = ligneous.AddLogger("proxy", "debug", "logs")
	config            Config
	config_file       string
	print_version     bool
	cache, _              = tilecache.NewTileCache(tilecache.Config{Type: "mbtiles", Directory: "cache"})
	MaxConcurrentJobs int = DEFAULT_MAX_CONCURRENT_JOBS
)

// ProxyClient http client for server proxy tile layers.
var proxyClient = &http.Client{
	Timeout: time.Second * 30,
	// The default is 2, which is generally too low for our request concurrency
	// in this program, resulting in unboundded growth and eventual exhaustion
	// of all available ports. This should keep the number of detatched TIME_WAIT
	// sockets to a minimum that matches our concurrency configuration.
	Transport: &http.Transport{
		MaxIdleConnsPerHost: int(MaxConcurrentJobs),
		// http://craigwickesser.com/2015/01/golang-http-to-many-open-files/
		// Dial: (&net.Dialer{
		// 	Timeout: 5 * time.Second,
		// }).Dial,
		TLSHandshakeTimeout: 15 * time.Second,
		IdleConnTimeout:     60 * time.Second,
		// https://golang.org/pkg/net/http/#Transport
		// https://stackoverflow.com/questions/39813587/go-client-program-generates-a-lot-a-sockets-in-time-wait-state?utm_medium=organic&utm_source=google_rich_qa&utm_campaign=google_rich_qa
		MaxIdleConns: int(MaxConcurrentJobs),
	},
}

func RandomIntBetween(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func subDomain() string {
	subs := []string{"a", "b", "c"}
	n := RandomIntBetween(0, 3)
	return subs[n]
}

func fetch(url string) ([]byte, error) {
	// prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(err)
	}

	// Look like a web browser running leaflet
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
	resp, err := proxyClient.Do(req)

	// always close the response-body, even if content is not required
	defer resp.Body.Close()

	if nil != err {
		return nil, err
	}

	blob, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return []byte{}, err
	}

	logger.Debug(fmt.Sprintf("PROXY GET %v %v", url, resp.StatusCode))

	if 200 != resp.StatusCode {
		err := errors.New("Request error: " + string(blob))
		return []byte{}, err
	}

	return blob, err
}

func fetchTile(layerUrl string, z, x, y uint64) ([]byte, error) {
	// build url
	tileUrl := strings.Replace(layerUrl, "{z}", fmt.Sprintf("%v", z), -1)
	tileUrl = strings.Replace(tileUrl, "{x}", fmt.Sprintf("%v", x), -1)
	tileUrl = strings.Replace(tileUrl, "{y}", fmt.Sprintf("%v", y), -1)
	tileUrl = strings.Replace(tileUrl, "{s}", subDomain(), -1)

	// Retry attempts -- 5
	attempt := 0
	for {
		// fetch data
		blob, err := fetch(tileUrl)

		// retry if there is an error
		if nil != err {
			attempt++
			if attempt > 4 {
				return []byte{}, err
			}
		}

		// return data
		return blob, err
	}
}

func jsonHttpResponse(w http.ResponseWriter, status_code int, result string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status_code)
	var payload string
	if 300 <= status_code {
		payload = fmt.Sprintf(`{"status":"error","error": "%v"}`, result)
		logger.Error(payload)
	} else {
		payload = fmt.Sprintf(`{"status":"ok","data": %v}`, result)
		if DEBUG {
			logger.Debug(payload)
		}
	}
	fmt.Fprintln(w, payload)
}

func GetTileHandler(layers map[string]string, contentType) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		z, _ := strconv.ParseUint(vars["z"], 10, 64)
		x, _ := strconv.ParseUint(vars["x"], 10, 64)
		y, _ := strconv.ParseUint(vars["y"], 10, 64)

		layer_name := strings.ToLower(vars["layer_name"])

		if _, ok := layers[layer_name]; ok {

			tile, err := cache.GetTile(layer_name, uint32(x), uint32(y), uint32(z))
			if nil == err {
				logger.Debug(fmt.Sprintf("CACHE GET %v %v %v %v", layer_name, z, x, y))
				w.Header().Set("Content-Type", contentType)
				w.WriteHeader(http.StatusOK)
				_, err = w.Write(tile)
				if err != nil {
					logger.Error(err)
				}
				return
			}

			tile, err = fetchTile(layers[layer_name], z, x, y)
			if nil != err {
				err := errors.New("Unable to fetch tile")
				jsonHttpResponse(w, 500, err.Error())
				return
			}

			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(tile)
			if err != nil {
				logger.Error(err)
			}

			err = cache.SetTile(layer_name, uint32(x), uint32(y), uint32(z), tile)
			if err != nil {
				logger.Error(err)
			}

			return
		}

		err := errors.New("Layer not found")
		jsonHttpResponse(w, 404, err.Error())

	}

}

func NewRouter(layers map[string]string) *mux.Router {
	router := mux.NewRouter()
	// router.HandleFunc("/api/v1/layers", GetLayersHandler(layers)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/layer/{layer_name}/tile/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}", GetTileHandler(layers, "image/png")).Methods("GET", "OPTIONS")

	return router
}

func init() {
	flag.StringVar(&config_file, "c", "", "tile server config")
	flag.BoolVar(&print_version, "v", false, "version")
	flag.Parse()
	if print_version {
		fmt.Println("MBTile Proxy 0.0.1")
		os.Exit(1)
	}
}

func getConfig(file string) {
	// check if file exists!!!
	if _, err := os.Stat(file); err == nil {

		fileHandler, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(fileHandler, &config)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		logger.Debug(config)
	} else {
		fmt.Println("Config file not found")
		os.Exit(1)
	}
}

// Before uncommenting the GenerateOSMTiles call make sure you have
// the necessary OSM sources. Consult OSM wiki for details.
func main() {
	getConfig(config_file)

	bind := fmt.Sprintf("0.0.0.0:%v", config.Port)
	router := NewRouter(config.Layers)
	logger.Info(fmt.Sprintf("Magic happens on port %v...", config.Port))
	srv := &http.Server{
		Addr:         bind,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	logger.Error(srv.ListenAndServe())
}
