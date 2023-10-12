//
// Copyright (C) 2020 OpenSIPS Solutions
//
// Call API is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Call API is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/OpenSIPS/call-api/internal/jsonrpc"
	"github.com/OpenSIPS/call-api/pkg/config"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func usage(prog string) {
	logrus.Fatalf("Usage: %s jsonrpc_method [jsonrpc_arguments]", prog)
}

func ParseClientArgs() (string, string, string, interface{}, string) {
	var wsServer, method, params, id string

	flag.StringVar(&wsServer, "wsserver", "localhost", "The API host to connect to")
	flag.StringVar(&method, "method", "", "JSON-RPC method")
	flag.StringVar(&params, "params", "", "JSON-RPC params")
	flag.StringVar(&id, "id", "", "JSON-RPC id")

	cfgPath, err := config.ParseFlags("call-api")
	if err != nil {
		logrus.Fatal(err)
	}

	if method == "" {
		logrus.Error("no method specified!")
		usage(os.Args[0])
	}

	var v interface{}

	if params != "" {
		err = json.Unmarshal([]byte(params), &v)
		if err != nil {
			logrus.Fatalf("failed to parse JSON args: %s", err)
		}
	}

	return wsServer, cfgPath, method, v, id
}

func closeWSConnection(c *websocket.Conn) {

	logrus.Info("gracefully closing connection...")

	// Cleanly close the connection by sending a close message
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		logrus.Errorf("write close: %s", err)
		return
	}
}

func main() {
	// parse cmdline args
	wsServer, cfgPath, method, params, id := ParseClientArgs()

	// read configuration
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		logrus.Fatal(err)
	}

	// prepare logging
	logfile, err := config.InitLogging(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	if logfile != nil {
		defer logfile.Close()
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	api_hostport := fmt.Sprintf("%s:%d", wsServer, cfg.WSServer.Port)
	u := url.URL{Scheme: "ws", Host: api_hostport, Path: cfg.WSServer.Path}
	logrus.Printf("connecting to %s", u.String())

	// open a single WebSocket connection
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// keep listening for messages until EOF
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logrus.Println("read:", err)
				return
			}

			logrus.Printf("recv: %s", message)

			var v interface{}

			err = json.Unmarshal(message, &v)
			if err != nil {
				logrus.Println("failed to parse JSON reply: %s", err)
				return
			}

			params := v.(map[string]interface{})["params"]
			if params != nil {
				status := params.(map[string]interface{})["event"]
				if status == "Ended" || status == "Error" {
					closeWSConnection(c)
					return
				}
			}
		}
	}()

	// create a JSON-RPC request
	req := jsonrpc.NewRequest(id, method, params)
	if req == nil {
		logrus.Fatal("failed to create JSON-RPC request")
	}

	// ... serialize it
	buf, err := req.Buffer()
	if err != nil {
		logrus.Fatal("write:", err)
	}
	logrus.Infof("send: %s", buf)

	// ... and send it!
	err = c.WriteMessage(websocket.TextMessage, buf)
	if err != nil {
		logrus.Fatal("write:", err)
	}

	select {
	case <-done:
		return
	case <-interrupt:
		logrus.Println("interrupt")
		closeWSConnection(c)

		// wait (with timeout) for the server to close the connection
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}

}
