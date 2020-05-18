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
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
)

type MIDatagram struct {
	conn *net.UDPConn
	buffer []byte
	idLock sync.Mutex
	id     uint64
	done   chan error
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

	r, _,  err := mi.conn.ReadFrom(mi.buffer)
	if err != nil {
		mi.done <- err
		return
	}

	reply := &jsonrpc.JsonRPCResponse{}
	err = reply.Parse(mi.buffer[0:r])
	if err != nil {
		mi.done <- err
		return
	}

	if reply.ID != currentId {
		mi.done <- errors.New("id mismatch")
		return
	}

	fn(reply, params)
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

	js := jsonrpc.NewRequest(currentId, command, params)
	jb, err := js.Buffer()
	if err != nil {
		return err
	}

	/* make sure the channel is drained */
	for len(mi.done) > 0 {
		<-mi.done
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

func (mi *MIDatagram) callSyncStore(response *jsonrpc.JsonRPCResponse, param interface{}) {

	var ok bool
	var replyChan chan *jsonrpc.JsonRPCResponse

	if replyChan, ok = param.(chan *jsonrpc.JsonRPCResponse); !ok {
		mi.done <- errors.New("invalid parameter passed at callback")
		return
	}

	replyChan <- response
}

func (mi *MIDatagram) CallSync(command string, param interface{}) (*jsonrpc.JsonRPCResponse, error) {
	replyChan := make(chan *jsonrpc.JsonRPCResponse, 1)

	err := mi.Call(command, param, mi.callSyncStore, replyChan);
	if err != nil {
		return nil, err
	}

	err = mi.Wait()
	if err != nil {
		return nil, err
	}

	return <- replyChan, nil
}
