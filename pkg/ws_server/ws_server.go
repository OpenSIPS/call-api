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

package ws_server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/call-api/internal/jsonrpc"
	"github.com/OpenSIPS/call-api/pkg/cmd"
	"github.com/OpenSIPS/call-api/pkg/config"
	"github.com/OpenSIPS/call-api/pkg/proxy"
)

const default_ws_host string = "localhost"
const default_ws_port int = 5059
const default_ws_path string = "/call-api"

var upgrader = websocket.Upgrader{} // use default options
var Cfg *config.Config

type WSConnection struct {
	conn *websocket.Conn
	proxy *proxy.Proxy // two-way UDP connection to a SIP proxy
}

type WSCmdEvent struct {
	cmd *cmd.Cmd
	event *cmd.CmdEvent
}

func (wsc *WSConnection) ReplyError(error_msg string, jsonrpc_id interface{}) {
	response := &jsonrpc.JsonRPCResponse{
		JSONRPC: "2.0",
		ID: jsonrpc_id,
		Error: &jsonrpc.JsonRPCError{
			Code: 32000,
			Message: error_msg,
		},
	}

	message, err := json.Marshal(response)
	if err != nil {
		logrus.Error("failed to build JSON-RPC error")
		return
	}

	err = wsc.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		logrus.Error("write: ", err)
	}
}

func (wsc *WSConnection) ReplyOK(jsonrpc_id interface{}, cmd_id string) {
	response := &jsonrpc.JsonRPCResponse{
		JSONRPC: "2.0",
		ID: jsonrpc_id,
		Result: &map[string]string {
			"status": "Started",
			"cmd_id": cmd_id,
		},
	}

	message, err := json.Marshal(response)
	if err != nil {
		logrus.Error("failed to build JSON-RPC error message")
		return
	}

	err = wsc.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		logrus.Error("write: ", err)
	}
}

// wait for random OpenSIPS MI events on a given WebSocket connection,
// possibly from multiple Call Commands running concurrently, and forward them
// to the WebSocket client as JSON-RPC Notifications
func (wsc *WSConnection) pollWSConnection(agg chan *WSCmdEvent) {
	var response *jsonrpc.JsonRPCNotification

	for ev := range agg {
		c := ev.cmd

		if ev.event == nil {
			response = &jsonrpc.JsonRPCNotification{
				JSONRPC: "2.0",
				Method: "Ended",
				Params: &map[string]interface{}{
					"cmd_id": c.ID,
				},
			}
		} else {
			logrus.Debugf("event on cmd %s (%s), event: %s", c.Command, c.ID, ev.event)

			if ev.event.IsError() {
				response = &jsonrpc.JsonRPCNotification{
					JSONRPC: "2.0",
					Method: "Error",
					Params: &map[string]interface{}{
						"cmd_id": c.ID,
						"error_msg": fmt.Sprintf("%s", ev.event.Error),
					},
				}
			} else {
				response = &jsonrpc.JsonRPCNotification{
					JSONRPC: "2.0",
					Method: "Event",
					Params: &map[string]interface{}{
						"cmd_id": c.ID,
						"data": ev.event.Event,
					},
				}
			}
		}

		message, err := json.Marshal(response)
		if err != nil {
			logrus.Errorf("cmd %s (%s): failed to build JSON notification: %s",
						  c.Command, c.ID, ev.event)
			return
		}

		err = wsc.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			logrus.Error("write: ", err)
		}
	}
}

// JSON-RPC based, two-way communication over a long-lived WebSocket connection
func wsConnection(w http.ResponseWriter, r *http.Request) {
	var err error
	var cmd_id string

	logrus.Debugf("new connection from %s", r.RemoteAddr)

	wsc := &WSConnection{}
	wsc.conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Print("upgrade:", err)
		return
	}
	defer wsc.conn.Close()

	logrus.Debug("upgraded to WebSocket")

	wsc.proxy = proxy.NewProxy(Cfg)
	if wsc.proxy == nil {
		logrus.Fatal("could not initialize SIP proxy")
	}

	agg := make(chan *WSCmdEvent)

	go wsc.pollWSConnection(agg)

	for {
		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			logrus.Info("read: ", err)
			break
		}

		logrus.Infof("recv: %s", message)

		// validate the incoming JSON-RPC query
		req := &jsonrpc.JsonRPCRequest{}
		err = req.Parse(message)
		if err != nil {
			wsc.ReplyError("failed to parse JSON-RPC request", "")
			continue
		}

		params, ok := req.Params.(map[string]interface{})
		if !ok {
			wsc.ReplyError("non-object parameters are not accepted", req.ID)
			continue
		}

		cmd_any_id, ok := params["cmd_id"]
		// if there was a cmd_id in the initial request
		if ok {
			cmd_id, ok = cmd_any_id.(string)
			if !ok {
				wsc.ReplyError("bad cmd_id (must be a string)", req.ID)
				continue
			}
		} else {
			cmd_id = ""
		}

		c := cmd.New(req.Method, cmd_id, wsc.proxy)
		if c == nil {
			wsc.ReplyError("unknown JSON-RPC method", req.ID)
			continue
		}

		// we expect to receive at least a close on this command's channel
		go func(c *cmd.Cmd) {
			for event := range c.Wait() {
				agg <- &WSCmdEvent{c, event}
			}

			logrus.Debugf("done reading events for cmd %s (%s)", c.Command, c.ID)
			agg <- &WSCmdEvent{c, nil}
		}(c)

		// launch the Calling command to run asynchronously
		err = c.Run(params)
		if err != nil {
			wsc.ReplyError("bad JSON-RPC parameters", req.ID)
			continue
		}

		// indicate that we've successfully launched the command
		wsc.ReplyOK(req.ID, c.ID)
	}

	close(agg)
	logrus.Debugf("closed connection from %s", r.RemoteAddr)
}

func Run(cfg *config.Config) {
	var host, path string
	var port int
	Cfg = cfg

	if cfg.WSServer.Host != "" {
		host = cfg.WSServer.Host
	} else {
		host = default_ws_host
	}

	if cfg.WSServer.Port != 0 {
		port = cfg.WSServer.Port
	} else {
		port = default_ws_port
	}

	if cfg.WSServer.Path != "" {
		path = cfg.WSServer.Path
	} else {
		path = default_ws_path
	}

	http.HandleFunc(path, wsConnection)

	listen := fmt.Sprintf("%s:%d", host, port)
	logrus.Infof("Listening for JSON-RPC over WebSocket on %s%s ...", listen, path)
	logrus.Fatal(http.ListenAndServe(listen, nil))
}
