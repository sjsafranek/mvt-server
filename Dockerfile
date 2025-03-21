FROM golang:1.12
#FROM debian:9

RUN apt-get update && apt-get install -y wget git postgresql

# RUN wget -q https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
# RUN tar -C /usr/local -xzf go1.11.2.linux-amd64.tar.gz
# RUN export PATH=$PATH:/usr/local/go/bin

RUN go get github.com/gorilla/mux
RUN go get github.com/lib/pq
RUN go get github.com/paulmach/orb
RUN go get github.com/pelletier/go-toml
RUN go get github.com/sjsafranek/goutils
RUN go get github.com/sjsafranek/ligneous
RUN go get github.com/garyburd/redigo/redis
RUN go get github.com/mattn/go-sqlite3

#RUN git clone https://github.com/sjsafranek/mvt-server.git && \
#    cd mvt-server && \
#    go run *.go -h db -dbp 1111

COPY . /go/src/mvt-server

CMD cd /go/src/mvt-server && \
    ./wait-for-postgres.sh db && \
    cd /go/src/mvt-server && \
    go run *.go -h db -dbp 5432

EXPOSE 5555
