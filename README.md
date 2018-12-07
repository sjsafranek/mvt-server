# mvt-server
Mapbox Vector Tile server written in Go and PostGIS

## Requires
PostGIS 2.5
PostGreSQL 11
Go 11





libprotobuf-c-dev
libprotobuf-dev
libprotoc-dev
protobuf-c-compiler
libprotobuf-c1





apt-get install libprotobuf-c-dev
apt-get install libprotobuf-c1
apt-get install libprotoc-dev
apt-get install protobuf-c-compiler
apt-get install libprotoc-dev




./mvt-server -action upload data/USA-NewYorkCity.shp 'usa-newyorkcity-12-04-2018' 'usa nyc roads' 4269


go get github.com/gorilla/mux
go get github.com/lib/pq
go get github.com/paulmach/orb
go get github.com/pelletier/go-toml
go get github.com/sjsafranek/goutils
go get github.com/sjsafranek/ligneous


./mvt-server -action upload ~/DB4IoT/tiger_shapes/tl_2017_us_county/tl_2017_us_county.shp tl_2017_us_county 'tiger line counties' 4269












# DOCKER DB

docker run --name mvt-postgis -e POSTGRES_PASSWORD=dev -d mdillon/postgis

docker run -it --rm --link mvt-postgis:postgres postgres psql -h postgres -U postgres

psql -h 192.168.16.2 -p 5432 -d postgres -U postgres

shp2pgsql -I data/USA-NewYorkCity.shp inrix-usa-newyorkcity-12-04-2018 | PGPASSWORD=dev psql -h 192.168.16.2 -p 5432 -d postgres -U postgres











docker-compose up
psql -h 127.0.0.1 -p 1111 -d geodev -U geodev











git clone https://github.com/sjsafranek/mvt-server.git




# Installing Docker

sudo apt install apt-transport-https ca-certificates curl gnupg2 software-properties-common
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
sudo aptitude update

sudo apt install docker-ce


sudo systemctl status docker
