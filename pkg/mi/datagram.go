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
	"io"
	"log"
	"errors"
	"net/rpc"
	"github.com/powerman/rpc-codec/jsonrpc2"

)

type MIDatagram struct {
	conn *jsonrpc2.Client
}

func (mi *MIDatagram) Connect(url string) error {
	/* TODO: make a wiser detection here when/if we have multiple backends */
	conn, err := jsonrpc2.Dial("udp", url)
	if err != nil {
		return errors.New("cannot connect to udp:" + url)
	}
	mi.conn = conn
	return nil
}

func (mi *MIDatagram) Call(command string, param interface{}) (map[string]interface{}, error) {
	var reply map[string]interface{}

	err := mi.conn.Call(command, param, &reply)
	log.Printf("returned")
	if err == rpc.ErrShutdown || err == io.ErrUnexpectedEOF {
		return nil, errors.New("Connection error")
	} else if err != nil {
		rpcerr := jsonrpc2.ServerError(err)
		return nil, errors.New(rpcerr.Message)
	}
	return reply, nil
}
