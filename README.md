# mvt-server
Mapbox Vector Tile server written in Go and PostGIS

## Requires
PostGIS 2.5
PostGreSQL 11
Go 11


./mvt-server upload data/USA-NewYorkCity.shp 'usa-newyorkcity-12-04-2018' 'inrix usa nyc roads' 4269


go get github.com/gorilla/mux
go get github.com/lib/pq
go get github.com/paulmach/orb
go get github.com/pelletier/go-toml
go get github.com/sjsafranek/goutils/hashers
go get github.com/sjsafranek/goutils/shell
go get github.com/sjsafranek/ligneous





git clone https://github.com/sjsafranek/mvt-server.git




# Installing Docker

sudo apt install apt-transport-https ca-certificates curl gnupg2 software-properties-common
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
sudo aptitude update

sudo apt install docker-ce


sudo systemctl status docker
