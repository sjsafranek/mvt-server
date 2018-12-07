FROM debian:9

RUN apt-get update && apt-get install -y wget git sudo

RUN wget -q https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.11.2.linux-amd64.tar.gz
RUN export PATH=$PATH:/usr/local/go/bin

RUN /usr/local/go/bin/go get github.com/gorilla/mux
RUN /usr/local/go/bin/go get github.com/lib/pq
RUN /usr/local/go/bin/go get github.com/paulmach/orb
RUN /usr/local/go/bin/go get github.com/pelletier/go-toml
RUN /usr/local/go/bin/go get github.com/sjsafranek/goutils
RUN /usr/local/go/bin/go get github.com/sjsafranek/ligneous

RUN git clone https://github.com/sjsafranek/mvt-server.git

EXPOSE 5555
