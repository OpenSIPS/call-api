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

package ws_server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
	"github.com/OpenSIPS/opensips-calling-api/pkg/cmd"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/pkg/proxy"
)

const default_ws_host string = "localhost"
const default_ws_port int = 5059

var upgrader = websocket.Upgrader{} // use default options
var Cfg *config.Config

type WSConnection struct {
	conn *websocket.Conn
	proxy *proxy.Proxy // two-way UDP connection to a SIP proxy
}

func (wsc *WSConnection) ReplyError(error_msg string) {
	response := &jsonrpc.JsonRPCResponse{
		JSONRPC: "2.0",
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
		logrus.Println("write:", err)
	}
}

func (wsc *WSConnection) ReplyErrorID(error_msg string, jsonrpc_id interface{}) {
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
		logrus.Println("write:", err)
	}
}

func (wsc *WSConnection) ReplyOK(jsonrpc_id interface{}, cmd_id string) {
	response := &jsonrpc.JsonRPCResponse{
		JSONRPC: "2.0",
		ID: jsonrpc_id,
		Result: &map[string]string {
			"status": "ok",
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
		logrus.Println("write:", err)
	}
}

// JSON-RPC based, two-way communication over a long-lived WebSocket connection
func wsConnection(w http.ResponseWriter, r *http.Request) {
	var err error

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

	for {
		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			logrus.Println("read:", err)
			break
		}
		logrus.Printf("recv: %s", message)

		// validate the incoming JSON-RPC query
		req := &jsonrpc.JsonRPCRequest{}
		err = req.Parse(message)
		if err != nil {
			wsc.ReplyError("failed to parse JSON-RPC request")
			continue
		}

		params, ok := req.Params.(map[string]interface{})
		if !ok {
			wsc.ReplyErrorID("non-object parameters are not accepted", req.ID)
			continue
		}

		cmd_id, ok := params["cmd_id"].(string)
		if cmd_id != "" && !ok {
			wsc.ReplyErrorID("bad cmd_id: must be a string)", req.ID)
			continue
		}

		c := cmd.New(req.Method, cmd_id, wsc.proxy)
		if c == nil {
			wsc.ReplyErrorID("unknown JSON-RPC method", req.ID)
			continue
		}

		err = c.Run(params) // async
		if err != nil {
			wsc.ReplyErrorID("bad JSON-RPC parameters", req.ID)
			continue
		}

		// indicate that we've successfully launched the command
		wsc.ReplyOK(req.ID, c.ID)
	}
}

func Run(cfg *config.Config) {
	var host string
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

	http.HandleFunc("/ws", wsConnection)

	listen := fmt.Sprintf("%s:%d", host, port)
	logrus.Infof("Listening for JSON-RPC over WebSocket on %s ...", listen)
	logrus.Fatal(http.ListenAndServe(listen, nil))
}
