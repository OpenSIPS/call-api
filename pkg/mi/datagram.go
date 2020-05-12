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
	"errors"
	"encoding/json"
)

type MIDatagram struct {
	conn *net.UDPConn
	buffer []byte
	idLock sync.Mutex
	id     uint64
	done   chan error
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
	mi.done = make(chan error, 1)
	return nil
}

func (mi *MIDatagram) Addr() (net.Addr) {
	return mi.conn.RemoteAddr()
}

func (mi *MIDatagram) getReply(currentId uint64, fn MIreply, params interface{}) {

	var reply miResponse

	r, _,  err := mi.conn.ReadFrom(mi.buffer)
	if err != nil {
		mi.done <- err
		return
	}

	err = json.Unmarshal(mi.buffer[0:r], &reply)
	if err != nil {
		mi.done <- err
		return
	}

	if reply.ID != currentId {
		mi.done <- errors.New("id mismatch")
		return
	}

	if reply.Error != nil {
		mi.done <- errors.New(reply.Error.Message)
		return
	}
	if result, ok := reply.Result.(map[string]interface{}); ok {
		fn(result, params)
	} else {
		fn(nil, params)
	}
	mi.done <- nil
}

func (mi *MIDatagram) Wait() (error) {
	return <- mi.done
}

func (mi *MIDatagram) Call(command string, params interface{}, fn MIreply, fnp interface{}) (error) {

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
		Params:  params,
		JSONRPC: "2.0",
		ID:      currentId,
	}
	jb, err := json.Marshal(js)
	if err != nil {
		return err
	}
	/* writing the request */
	mi.conn.SetWriteDeadline(time.Now().Add(time.Second))
	_, err = mi.conn.Write(jb)
	if err != nil {
		return err
	}

	/* waiting for the reply */
	go mi.getReply(currentId, fn, fnp)
	return nil
}

/*
func (mi *MIDatagram) CallSync(command string, param interface{}) (map[string]interface{}, error) {
	return nil, nil
}
*/
