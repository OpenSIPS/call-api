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

package mi

import (
	"net"
	"time"
	"sync"
	"bufio"
	"errors"
	"encoding/json"
)

type MIDatagram struct {
	conn *net.UDPConn
	buffer []byte
	reader *bufio.Reader
	idLock sync.Mutex
	id     uint64
}

type miRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      uint64                 `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type miError struct {
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Data    interface{}             `json:"data,omitempty"`
}

type miResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      uint64                 `json:"id"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *miError               `json:"error,omitempty"`
}

func (mi *MIDatagram) Connect(url string) error {

	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	conn.SetReadBuffer(65535)
	conn.SetWriteBuffer(65535)
	mi.buffer = make([]byte, 65535)
	mi.conn = conn
	mi.reader = bufio.NewReader(conn)
	return nil
}

func (mi *MIDatagram) Call(command string, param interface{}) (map[string]interface{}, error) {

	var reply miResponse

	mi.idLock.Lock()
	currentId := mi.id
	mi.id += 1
	mi.idLock.Unlock()

	js := struct {
		Method  string           `json:"method"`
		Params  interface{}      `json:"params,omitempty"`
		JSONRPC string           `json:"jsonrpc"`
		ID      uint64           `json:"id"`
	}{
		Method:  command,
		Params:  param,
		JSONRPC: "2.0",
		ID:      currentId,
	}
	jb, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}
	/* writing the request */
	mi.conn.SetWriteDeadline(time.Now().Add(time.Second))
	_, err = mi.conn.Write(jb)
	if err != nil {
		return nil, err
	}

	/* waiting for the reply */
	r, err := mi.reader.Read(mi.buffer)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(mi.buffer[0:r], &reply)
	if err != nil {
		return nil, err
	}

	if reply.ID != currentId {
		return nil, errors.New("id mismatch")
	}

	if reply.Error != nil {
		return nil, errors.New(reply.Error.Message)
	}
	if result, ok := reply.Result.(map[string]interface{}); ok {
		return result, nil
	} else {
		return nil, nil
	}
}
