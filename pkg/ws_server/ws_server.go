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
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
)

var upgrader = websocket.Upgrader{} // use default options

// JSON-RPC based, two-way communication over a long-lived WebSocket connection
func wsConnection(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Print("upgrade:", err)
		return
	}
	defer c.Close()

	logrus.Debug("upgraded to WebSocket")

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			logrus.Println("read:", err)
			break
		}
		logrus.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			logrus.Println("write:", err)
			break
		}
	}
}

func Run(cfg *config.Config) {
	var host string
	var port int

	if cfg.WSServer.Host != "" {
		host = cfg.WSServer.Host
	} else {
		host = "localhost"
	}

	if cfg.WSServer.Port != 0 {
		port = cfg.WSServer.Port
	} else {
		port = 5059
	}

	http.HandleFunc("/ws", wsConnection)

	listen := fmt.Sprintf("%s:%d", host, port)
	logrus.Info("Listening for JSON-RPC over WebSocket on " + listen + " ...")
	logrus.Fatal(http.ListenAndServe(listen, nil))
}
