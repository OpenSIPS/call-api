//
// Copyright (c) 2020 OpenSIPS Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
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

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
)

func usage(prog string) {
	logrus.Fatalf("Usage: %s jsonrpc_method [jsonrpc_arguments]", prog)
}

func ParseClientArgs() (string, string, interface{}, string) {
	var method, params, id string

	flag.StringVar(&method, "method", "", "JSON-RPC method")
	flag.StringVar(&params, "params", "", "JSON-RPC params")
	flag.StringVar(&id, "id", "", "JSON-RPC id")

	cfgPath, err := config.ParseFlags("config")
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
			logrus.Fatalf("failed to parse JSON args: %s\n", err)
		}
	}

	return cfgPath, method, v, id
}

func main() {
	// parse cmdline args
	cfgPath, method, params, id := ParseClientArgs()

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

	// prepare the JSON-RPC data
	logrus.Debugf("cmd: %s, params: %s, id: %s\n", method, params, id)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	api_hostport := fmt.Sprintf("%s:%d", cfg.WSServer.Host, cfg.WSServer.Port)
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

		// Cleanly close the connection by sending a close message and then
		// waiting (with timeout) for the server to close the connection.
		err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			logrus.Println("write close:", err)
			return
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
		return
	}
}
