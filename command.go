package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"runtime"
	"strings"
	"time"
)

const (
	DEFAULT_TCP_PORT = 9622
	DEFAULT_TCP_HOST = "0.0.0.0"
)

var (
	TCP_PORT = DEFAULT_TCP_PORT
	TCP_HOST = DEFAULT_TCP_HOST
)

type CommandRequestMessage struct {
	Method      string `json:"method"`
	FilePath    string `json:"file_path,omitempty"`
	LayerName   string `json:"layer_name,omitempty"`
	Description string `json:"description,omitempty"`
	SRID        int64  `json:"srid,omitempty"`
}

func commandListener() {
	startTime := time.Now()

	var listener net.Listener
	c := 0
	for {
		serv := fmt.Sprintf("%v:%v", TCP_HOST, TCP_PORT)
		l, err := net.Listen("tcp", serv)
		if err != nil {
			logger.Warnf("Error listening: %v", err.Error())
			TCP_PORT++
			c++
			if c > 10 {
				logger.Errorf("Error listening: %v", err.Error())
				panic(err)
				return
			}
			continue
		}
		listener = l
		logger.Infof("Tcp Listening on %v", serv)
		break
	}

	// Close the listener when the application closes.
	defer listener.Close()

	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf("Error accepting connection: %v", err.Error())
			return
		}

		logger.Debugf("%v %v", conn.RemoteAddr().String(), "Connection open")

		// Handle connections in a new goroutine.
		go func(conn net.Conn) {
			defer conn.Close()
			defer logger.Warnf("%v %v", conn.RemoteAddr().String(), "Connection closed")

			reader := bufio.NewReader(conn)
			tp := textproto.NewReader(reader)

			for {

				// will listen for message to process ending in newline (\n)
				message, err := tp.ReadLine()
				if io.EOF == err {
					break
				}

				// No message was sent
				if "" == message {
					continue
				}

				// Command
				exitFlag := false
				switch {
				// case strings.HasPrefix(message, "help"):
				// 	response := self.Help()
				// 	self.HandleSuccess(response, conn)
				// 	continue
				case strings.HasPrefix(message, "quit"):
					fallthrough
				case strings.HasPrefix(message, "bye"):
					fallthrough
				case strings.HasPrefix(message, "exit"):
					exitFlag = true
				}
				if exitFlag {
					break
				}
				//.end

				logger.Debugf("%v Message Recieved: %v", conn.RemoteAddr().String(), string([]byte(message)))

				// json parse message
				req := CommandRequestMessage{}
				err = json.Unmarshal([]byte(message), &req)
				if err != nil {
					// invalid message
					// close connection
					// '\x04' end of transmittion character
					logger.Error("%v %v", conn.RemoteAddr().String(), err.Error())
					resp := `{"status": "error", "error": "` + fmt.Sprintf("%v", err) + `",""}`
					conn.Write([]byte(resp + "\n"))
					continue
				}

				switch {
				case "" == req.Method:
					// No method provided
					continue

				case "ping" == req.Method:
					// {"method":"ping"}
					conn.Write([]byte(`{"status":"ok","data":{"message":"pong"}}` + "\n"))
					continue

				case "get_runtime_stats" == req.Method:
					// {"method":"get_runtime_stats"}
					var ms runtime.MemStats
					runtime.ReadMemStats(&ms)
					response := fmt.Sprintf(`{"status":"ok","data":{"num_goroutine":%v,"alloc":%v,"total_alloc":%v,"sys":%v,"NumGC":%v,"registered":"%v","uptime":%v,"num_cpu":%v,"goos":"%v"}}`,
						runtime.NumGoroutine(), ms.Alloc/1024, ms.TotalAlloc/1024, ms.Sys/1024, ms.NumGC, startTime.UTC(), time.Since(startTime).Seconds(), runtime.NumCPU(), runtime.GOOS)

					conn.Write([]byte(response + "\n"))
					continue

				case "upload" == req.Method:
					// 	{"method":"upload","file_path": "/home/stefan/Repos/mvt-server/data/inrix_shapefiles/copenhagen/Oresundsanalys_draft_180807.shp", "layer_name":"test_layer-12-13-2018", "description":"roads of denmark inrix", "srid": 4326}
					_, err := UploadShapefile(req.FilePath, req.LayerName, req.Description, req.SRID)
					if err != nil {
						logger.Error(err)
						conn.Write([]byte(fmt.Sprintf(`{"status":"error","error":{"message":"%v"}}`, err.Error()) + "\n"))
						continue
					}

					err = LAYERS.AddLayer(req.LayerName)
					if err != nil {
						logger.Error(err)
						conn.Write([]byte(fmt.Sprintf(`{"status":"error","error":{"message":"%v"}}`, err.Error()) + "\n"))
						continue
					}

					conn.Write([]byte(`{"status":"ok","data":{"message":"layer created"}}` + "\n"))
					continue

				case "delete" == req.Method:
					// {"method":"delete", "layer_name":"test_layer-12-13-2018"}
					err := LAYERS.DeleteLayer(req.LayerName)
					if err != nil {
						logger.Error(err)
						conn.Write([]byte(fmt.Sprintf(`{"status":"error","error":{"message":"%v"}}`, err.Error()) + "\n"))
						continue
					}

					conn.Write([]byte(`{"status":"ok","data":{"message":"layer deleted"}}` + "\n"))
					continue

				case "panic" == req.Method:
					// {"method":"panic"}
					// os.Exit(1)
					panic(errors.New("Panic"))
					continue

				}

			}

		}(conn)

	}
}
