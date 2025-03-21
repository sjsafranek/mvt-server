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
github.com/garyburd/redigo/redis
```

### Exporting from PostGIS
pgsql2shp -f <path to output shapefile> -h <hostname> -u <username> -P <password> databasename "<query>"
pgsql2shp -f tl_2018_us_place.shp inrix "select * from tl_2018_us_place;"
pgsql2shp -f or_segments_20190401.shp inrix "select * from trajectory_segments_view;"











{"method":"upload","file_path": "/home/stefan/go/src/mvt-server/data/ne_10m_admin_1_states_provinces/copenhagen/ne_10m_admin_1_states_provinces.shp", "layer_name":"ne_10m_admin_1_states_provinces", "description":"natural earth data", "srid": 4326}

{"method":"upload","file_path": "/home/stefan/Repos/mvt-server/data/inrix_shapefiles/copenhagen/Oresundsanalys_draft_180807.shp", "layer_name":"test_layer-12-13-2018", "description":"roads of denmark inrix", "srid": 4326}

{"method":"upload","file_path": "/home/stefan/mvt-server/data/inrix_netherlands/Uitvoer_shape/wijk_2018.shp", "layer_name":"uitvoer_wijk_2018", "description":"Uitvoer wijk", "srid": 28992}

{"method":"upload","file_path": "/home/stefan/mvt-server/data/inrix_netherlands/Uitvoer_shape/gem_2018.shp", "layer_name":"uitvoer_gem_2018", "description":"Uitvoer gem", "srid": 28992}

{"method":"upload","file_path": "/home/stefan/mvt-server/data/inrix_netherlands/Uitvoer_shape/buurt2018.shp", "layer_name":"uitvoer_buurt_2018", "description":"Uitvoer buurt", "srid": 28992}


{"method":"delete", "layer_name":"test_layer-12-13-2018"}


redis-cli KEYS "inrix*" | xargs redis-cli DEL






./mvt-server_00.01.07 -action cook clark_county_segments '{"conditions":[{"test":"match","column_id":"frc","values":["0","1","2"]}]}' 11 12 45.39555704145539 45.84123776445225 -122.82714843750001 -122.36297607421876

./mvt-server_00.01.07 -action cook clark_county_segments '' 14 16 45.39555704145539 45.84123776445225 -122.82714843750001 -122.36297607421876












{"method":"upload","file_path": "/home/stefan/shapefiles/munich/Muenchen_Stadtviertel.shp", "layer_name":"stadtviertel_muenchen", "description":"stadtviertel_muenchen", "srid": 4326}








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
TileCache struct



sudo docker inspect mvt-server_db_1_1517c26d31ab
"Gateway": "172.21.0.1"


```bash


SELECT DISTINCT ST_GeometryType(geom) FROM "inrix_england_and_wales_12-12-2018";


ALTER TABLE "tl_2017_us_zcta510"
    ALTER COLUMN geom TYPE geometry(MultiPolygon,3857)
    USING ST_Transform(
        ST_SetSRID( geom, 4269 )
        , 3857
    );

UPDATE LAYERS SET srid=3857 WHERE layer_name = 'tl_2017_us_zcta510';


```
