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
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
)

func main() {
	// parse cmdline args
	cfgPath, err := config.ParseFlags("config")
	if err != nil {
		logrus.Fatal(err)
	}

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

	listen := fmt.Sprintf("%s:%d", cfg.WSServer.Host, cfg.WSServer.Port)
	u := url.URL{Scheme: "ws", Host: listen, Path: "/ws"}
	logrus.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

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

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				logrus.Println("write:", err)
				return
			}
		case <-interrupt:
			logrus.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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
}
