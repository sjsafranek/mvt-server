package main

import (
	"encoding/json"
	"errors"
	"fmt"
	// "html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func layerNotFoundHttpError(w http.ResponseWriter) {
	err := errors.New("Layer not found")
	jsonHttpResponse(w, 404, err.Error())
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

func formatFilter(filter string, layer *Layer) (string, error) {
	if "" != filter {
		// TODO:
		//  - Handle json filters...
		if strings.HasPrefix(filter, "{") && strings.HasSuffix(filter, "}") {
			var filters Filters
			err := json.Unmarshal([]byte(filter), &filters)
			if nil != err {
				return "", err
			}
			sql_filter, err := filters.ToSQL(layer)
			if nil != err {
				return "", err
			}
			filter = sql_filter
			return filter, nil
		}
	}
	return "", nil
}

func getFiltersFromRequest(r *http.Request, layer *Layer) (string, error) {
	filters, ok := r.URL.Query()["filter"]
	if ok {
		return formatFilter(filters[0], layer)
	}

	filters, ok = r.URL.Query()["filters"]
	if ok {
		return formatFilter(filters[0], layer)
	}

	return "", nil
}

func ApiV1TileMapServiceGetVectorTile(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	ctx := r.Context()

	select {
	case <-ctx.Done():
		logger.Tracef("request canceled %v", time.Since(start))

	default:
		// err := func() (error) {}()
		vars := mux.Vars(r)
		z, _ := strconv.ParseUint(vars["z"], 10, 64)
		x, _ := strconv.ParseUint(vars["x"], 10, 64)
		y, _ := strconv.ParseUint(vars["y"], 10, 64)

		layer_name := strings.ToLower(vars["layer_name"])

		layer, err := LAYERS.GetLayer(layer_name)
		if nil != err {
			layerNotFoundHttpError(w)
			return
		}

		filter, err := getFiltersFromRequest(r, layer)
		if nil != err {
			jsonHttpResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tileData, err := layer.FetchTileWithContext(ctx, uint32(x), uint32(y), uint32(z), filter)
		if nil != err {
			if nil != ctx.Err() && "context canceled" != ctx.Err().Error() {
				jsonHttpResponse(w, http.StatusInternalServerError, err.Error())
			} else {
				jsonHttpResponse(w, http.StatusRequestTimeout, err.Error())
			}
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(tileData)

		logger.Tracef("request completed %v", time.Since(start))

	}
}

func ApiV1LayersHandler(w http.ResponseWriter, r *http.Request) {
	b, err := LAYERS.MarshalToJSON()
	if nil != err {
		jsonHttpResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonHttpResponse(w, http.StatusOK, fmt.Sprintf(`{"layers": %v}`, string(b)))
}

func ApiV1LayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	layer_name := strings.ToLower(vars["layer_name"])

	layer, err := LAYERS.GetLayer(layer_name)
	if nil != err {
		layerNotFoundHttpError(w)
		return
	}

	b, err := layer.MarshalToJSON()
	if nil != err {
		jsonHttpResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonHttpResponse(w, http.StatusOK, fmt.Sprintf(`{"layer": %v}`, string(b)))
}

func ApiV1WebFeatureServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	layer_name := strings.ToLower(vars["layer_name"])
	layer, err := LAYERS.GetLayer(layer_name)
	if nil != err {
		layerNotFoundHttpError(w)
		return
	}

	decoder := json.NewDecoder(r.Body)
	query := FeatureServiceQuery{ToSRID: 4269} // default value
	err = decoder.Decode(&query)
	if err != nil {
		jsonHttpResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	sql_query, err := query.ToSQL()
	if err != nil {
		layerNotFoundHttpError(w)
		return
	}

	result, err := layer.QueryRowJSON(sql_query)
	if err != nil {
		jsonHttpResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonHttpResponse(w, http.StatusOK, fmt.Sprintf(`{"geojson": %v}`, result))
}

func ViewMapHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, `
<!DOCTYPE html>
<html>
	<head>

		<title>MVT Server</title>

		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0">

		<link rel="shortcut icon" type="image/x-icon" href="docs/images/favicon.ico" />

		<link rel="stylesheet" href="https://unpkg.com/leaflet@1.3.4/dist/leaflet.css" integrity="sha512-puBpdR0798OZvTTbP4A8Ix/l+A4dHDD0DGqYW6RQ+9jxkRFclaxxQb/SJAWZfWAkuyeQUytO7+7N4QKrDh+drA==" crossorigin=""/>
		<script src="https://unpkg.com/leaflet@1.3.4/dist/leaflet.js" integrity="sha512-nMMmRyTVoLYqjP9hrbed9S+FzjZHW5gY1TWCHA5ckwXZBadntCNs8kEqAWdrb9O7rxbCaA4lKTIWjDXZxflOcA==" crossorigin=""></script>

		<!-- <script src="https://unpkg.com/leaflet.vectorgrid@latest/dist/Leaflet.VectorGrid.bundled.js"></script> -->
		<script src="/static/leaflet_vectorlayer.js"></script>

		<script
		  src="https://code.jquery.com/jquery-3.3.1.min.js"
		  integrity="sha256-FgpCb/KJQlLNfOu91ta32o/NMZxltwRo8QtmkMRdAu8="
		  crossorigin="anonymous"></script>

		<style>
			body {
				padding: 0;
				margin: 0;
			}
			html, body, #mapid {
				height: 100%;
				width: 100%;
			}
		</style>

	</head>
	<body>

		<div id="mapid"></div>

		<script>

			var map;
			var overlayMaps = {};

			$(document).ready(function(){
				map = L.map('mapid').setView([0, 0], 2);

				L.tileLayer('https://api.tiles.mapbox.com/v4/{id}/{z}/{x}/{y}.png?access_token=pk.eyJ1IjoibWFwYm94IiwiYSI6ImNpejY4NXVycTA2emYycXBndHRqcmZ3N3gifQ.rJcFIG214AriISLbB6B5aw', {
					maxZoom: 18,
					attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, ' +
						'<a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, ' +
						'Imagery © <a href="https://www.mapbox.com/">Mapbox</a>',
					id: 'mapbox.streets'
				}).addTo(map);

				function newMvtLayer(info) {
					var lyrId = info.layer_name;

					var layer = L.vectorGrid.protobuf("/api/v1/layer/"+lyrId+"/tile/{z}/{x}/{y}.mvt?filters={filters}", {
						rendererFactory: L.canvas.tile,
						filters: "",
						interactive: true,
						getFeatureId: function(feature) {
							return JSON.stringify(feature.properties);
						},
						vectorTileLayerStyles: {
							layer: function(properties, zoom, geometryDimension) {
								return {
									fill: true,
									fillOpacity: 0,
									weight: 1,
									color: "black"
								};
							}
						}
					});

					// highlight feature on hover
					layer.on("mouseover", function(event) {
						console.log(event.layer.properties);

						for (var _id in this._overriddenStyles) {
							this.resetFeatureStyle(_id);
						}

						var _id = JSON.stringify(event.layer.properties);
						this.setFeatureStyle(_id, {
							fill: true,
							fillOpacity: 0,
							weight: 4,
							color: "black"
						});
					});

					// fetch feature on click
					layer.on("click", function(event) {
						console.log(event.layer.properties);

						var payload = {
							"method": "get_feature",
							"limit": 1,
							"layer": lyrId,
							"filters": [
								{
									"test": "contains",
									"wkt": "POINT("+event.latlng.lng+" "+event.latlng.lat+")"
								}
							]
						};

						$.ajax({
							type: "POST",
							url: "/api/v1/layer/"+lyrId+"/wfs",
							data: JSON.stringify(payload),
							contentType: "application/json"
							// contentType: "text/plain"
						})
						.done(function(data) {
							console.log(data);
						})
						.fail(function(data) {
							console.log(data);
						});

					});

					return layer;
				}

				$.getJSON('/api/v1/layers', function(data) {
					var layers = data.data.layers;
					for (var i=0; i<layers.length; i++) {
						var lyrId = layers[i].layer_name;
						var layer = newMvtLayer(layers[i]);
						overlayMaps[lyrId] = layer;
						overlayMaps[lyrId + '::extent'] = L.geoJSON(layers[i].extent);
					}

					var layerControl = L.control.layers({}, overlayMaps);
					layerControl.addTo(map);
				});

			});

		</script>

	</body>
</html>
`)
}
