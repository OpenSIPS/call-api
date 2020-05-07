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

package event

import (
	"net"
	"bufio"
	"errors"
	"syscall"
	"github.com/OpenSIPS/opensips-calling-api/pkg/mi"
)

type EventDatagram struct {
	mi mi.MI
	conn *net.UDPConn
	buffer []byte
	reader *bufio.Reader
}

func (event *EventDatagram) Connect(mi mi.MI) (error) {

	// we first need to check how we can connect to the MI handler
	miAddr, ok := mi.Addr().(*net.UDPAddr)
	if ok != true {
		return errors.New("using a different protocol to conect to MI")
	}
	c, err := net.DialUDP("udp", nil, miAddr)
	if err != nil {
		return err
	}

	udpAddr, ok := c.LocalAddr().(*net.UDPAddr)
	if ok != true {
		return errors.New("using a different protocol to conect to MI")
	}

	// we've now got the IP we can use to reach MI, use it for further events
	local := net.UDPAddr{IP: udpAddr.IP}
	conn, err := net.ListenUDP(c.LocalAddr().Network(), &local)
	if err != nil {
		return err
	}
	conn.SetReadBuffer(65535)
	conn.SetWriteBuffer(65535)

	file, _ := conn.File()
	fd := file.Fd()
	syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 0)

	event.buffer = make([]byte, 65535)
	event.conn = conn
	event.reader = bufio.NewReader(conn)
	return nil
}

func (event *EventDatagram) Socket() (string) {
	return event.conn.LocalAddr().String()
}
