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

package mi

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/OpenSIPS/call-api/internal/jsonrpc"
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

func (mi *MIDatagram) getReply(currentId uint64, fn MIreply) {

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

	replyId, ok := reply.ID.(float64)
	if !ok {
		mi.done <- errors.New("id type error")
		return
	}

	if uint64(replyId) != currentId {
		mi.done <- errors.New("id mismatch %d")
		return
	}

	if fn != nil {
		fn(reply)
	}
	mi.done <- nil
}

func (mi *MIDatagram) Wait() (error) {
	return <- mi.done
}

func (mi *MIDatagram) Call(command string, params interface{}, fn MIreply) (error) {

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
	go mi.getReply(currentId, fn)
	return nil
}

type miDatagramSync struct {
	reply chan *jsonrpc.JsonRPCResponse
}

func (mi *miDatagramSync) callSyncStore(response *jsonrpc.JsonRPCResponse) {

	mi.reply <- response
}

func (mi *MIDatagram) CallSync(command string, param interface{}) (*jsonrpc.JsonRPCResponse, error) {
	dgramSync := &miDatagramSync{make(chan *jsonrpc.JsonRPCResponse, 1)}

	err := mi.Call(command, param, dgramSync.callSyncStore);
	if err != nil {
		return nil, err
	}

	err = mi.Wait()
	if err != nil {
		return nil, err
	}

	return <- dgramSync.reply, nil
}
