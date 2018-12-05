FROM mdillon/postgis

RUN apt-get update && apt-get install -y wget git

RUN wget -q https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.11.2.linux-amd64.tar.gz
RUN export PATH=$PATH:/usr/local/go/bin

RUN /usr/local/go/bin/go get github.com/gorilla/mux
RUN /usr/local/go/bin/go get github.com/lib/pq
RUN /usr/local/go/bin/go get github.com/paulmach/orb
RUN /usr/local/go/bin/go get github.com/pelletier/go-toml
RUN /usr/local/go/bin/go get github.com/sjsafranek/goutils/hashers
RUN /usr/local/go/bin/go get github.com/sjsafranek/goutils/shell
RUN /usr/local/go/bin/go get github.com/sjsafranek/ligneous

RUN git clone https://github.com/sjsafranek/mvt-server.git
RUN psql -c "CREATE USER geodev WITH PASSWORD 'dev'"
RUN psql -c "CREATE DATABASE geodev"
RUN psql -c "GRANT ALL PRIVILEGES ON DATABASE geodev to geodev"
RUN psql -c "ALTER USER geodev WITH SUPERUSER"

RUN PGPASSWORD=dev psql -d geodev -U geodev -f mvt-server/scripts/database.sql

RUN cd mvt-server/ && /usr/local/go/bin/go run *.go

EXPOSE 5555
