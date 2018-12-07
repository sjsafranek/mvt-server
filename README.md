# mvt-server
Mapbox Vector Tile server written in Go and PostGIS

## Requires
 - PostGIS 2.5
 - PostGreSQL 11
 - Go 11

### MVT Tile Generation
 - libprotobuf-c-dev
 - libprotobuf-dev
 - libprotoc-dev
 - protobuf-c-compiler
 - libprotobuf-c1

### GoLang dependencies
```bash
go get github.com/gorilla/mux
go get github.com/lib/pq
go get github.com/paulmach/orb
go get github.com/pelletier/go-toml
go get github.com/sjsafranek/goutils
go get github.com/sjsafranek/ligneous
```


./mvt-server -action upload data/tl_2017_us_county/tl_2017_us_county.shp 'tl_2017_us_county' 'tiger line 2017 us counties' 4269
./mvt-server -action upload data/tl_2017_us_state/tl_2017_us_state.shp 'tl_2017_us_state' 'tiger line 2017 us states' 4269
./mvt-server -action upload data/tl_2017_us_zcta510/tl_2017_us_zcta510.shp 'tl_2017_us_zcta510' 'tiger line 2017 us zipcodes' 4269
./mvt-server -action upload data/United_Kingdom.shp 'inrix_united_kingdom_roads_12-04-2018' 'inrix uk road segments' 4269
./mvt-server -action upload data/ne_10m_admin_1_states_provinces/ne_10m_admin_1_states_provinces.shp 'ne_10m_admin_1_states_provinces' 'Natural Earth admin and provinces' 4269


## Docker Database Setup (optional)
```bash
docker-compose run -d

psql -h 127.0.0.1 -p 1111 -d geodev -U geodev
psql -p 1111 -h 0.0.0.0 -U geodev -d geodev
```

### Installing Docker (Debian)
```bash
sudo apt install apt-transport-https ca-certificates curl gnupg2 software-properties-common
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
sudo aptitude update

sudo apt install docker-ce
```


# TODO
Do srid ST_Transform to 3857 upon insertion



sudo docker inspect mvt-server_db_1_1517c26d31ab
"Gateway": "172.21.0.1"


```bash
ALTER TABLE "tl_2017_us_zcta510"
    ALTER COLUMN geom TYPE geometry(MultiPolygon,3857)
    USING ST_Transform(
        ST_SetSRID( geom, 4269 )
        , 3857
    );

UPDATE LAYERS SET srid=3857 WHERE layer_name = 'tl_2017_us_zcta510';
```
