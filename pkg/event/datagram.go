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
	"sync"
	"errors"
	"strings"
	"syscall"
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/mi"
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
)

// EventDatagram
type EventDatagramSub struct {
	event string
	conn *EventDatagramConn
	fn EventNotification
	fnp interface{}
}

func (sub *EventDatagramSub) String() (string) {
	return sub.conn.String()
}

func (sub *EventDatagramSub) Event() (string) {
	return sub.event
}

func (sub *EventDatagramSub) Unsubscribe() {
	logrus.Debug("unsubscribing event " + sub.event + " from " + sub.conn.String())
	sub.conn.Unsubscribe(sub)
}


// EventDatagramConn
type EventDatagramConn struct {
	udp *net.UDPConn
	wake chan error
	lock sync.RWMutex
	subs []*EventDatagramSub
	handler *EventDatagram
}

func (conn *EventDatagramConn) parse(result string) (*jsonrpc.JsonRPCNotification, error) {

	split := strings.Split(result, "\n")
	if len(split) < 1 || len(split[0]) < 1 {
		return nil, errors.New("no event specified")
	}
	event := split[0]
	ret := make(map[string]interface{})
	for _, line := range split[1:] {
		if len(line) == 0 {
			break
		}
		lineSplit := strings.Split(line, "::")
		if len(lineSplit) > 1 {
			ret[lineSplit[0]] = strings.Join(lineSplit[1:], "::")
		} else {
			ret[line] = nil
		}
	}

	return jsonrpc.NewNotification(event, ret), nil
}

func (conn *EventDatagramConn) waitForEvents() {

	buffer := make([]byte, 65535)
	for {
		select {
		case <-conn.wake:
			/* if there are no other subscribers, terminate the execution */
			conn.handler.lock.Lock()
			conn.lock.RLock()
			if len(conn.subs) == 0 {
				conn.lock.RUnlock()
				conn.handler.lock.Unlock()
				logrus.Info("closing connection " + conn.String())
				close(conn.wake)
				return
			}
			conn.lock.RUnlock()
			conn.handler.lock.Unlock()
		default:
			r, _, err := conn.udp.ReadFrom(buffer)
			if err == nil {
				result, err := conn.parse(string(buffer[0:r]))
				if err != nil {
					logrus.Error("could not parse notification: " + err.Error())
				} else {
					sub := conn.getSubscription(result.Method)
					// run in a different routine to avoid blocking
					if sub != nil {
						go sub.fn(sub, result, sub.fnp)
					} else {
						logrus.Warn("unknown subscriber for event " + result.Method)
					}
				}
			} else {
				logrus.Warn("error while listening for events: " + err.Error())
			}
		}
	}
}

func (conn *EventDatagramConn) Unsubscribe(sub *EventDatagramSub) {
	// first remove it from list, to make sure we don't get any other events
	// for it - locate it in the array
	conn.lock.Lock()
	for i, s := range conn.subs {
		if s == sub {
			conn.subs = append(conn.subs[0:i], conn.subs[i+1:]...)
			break
		}
	}
	conn.lock.Unlock()

	// inform the go routine it is no longer necessary to wait for events
	conn.wake <- nil

	// unsubscribe from the event
	/* we've got the connection - let us subscribe */
	var eviParams = map[string]interface{}{
		"event": sub.Event(),
		"socket": sub.String(),
		"expire": 0,
	}
	_, err := conn.handler.mi.CallSync("event_subscribe", &eviParams);
	if err != nil {
		logrus.Error("could not unsubscribe for event " + sub.Event() + " " + err.Error())
	}

}

func (conn *EventDatagramConn) Init(event *EventDatagram) (*EventDatagramConn) {

	// we first need to check how we can connect to the MI handler
	miAddr, ok := event.mi.Addr().(*net.UDPAddr)
	if ok != true {
		logrus.Error("using a different protocol to conect to MI")
		return nil
	}
	c, err := net.DialUDP("udp", nil, miAddr)
	if err != nil {
		logrus.Error(err)
		return nil
	}

	udpAddr, ok := c.LocalAddr().(*net.UDPAddr)
	if ok != true {
		logrus.Error("using a different protocol to conect to MI")
		return nil
	}

	// we've now got the IP we can use to reach MI, use it for further events
	local := net.UDPAddr{IP: udpAddr.IP}
	udpConn, err := net.ListenUDP(c.LocalAddr().Network(), &local)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	udpConn.SetReadBuffer(65535)
	udpConn.SetWriteBuffer(65535)

	file, _ := udpConn.File()
	fd := file.Fd()
	syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 0)

	/* we typically only have one subscriber per conn - lets start with that */
	conn.subs = make([]*EventDatagramSub, 0, 1)
	conn.wake = make(chan error, 1)
	conn.udp = udpConn
	conn.handler = event
	go conn.waitForEvents()
	return conn
}

func (conn *EventDatagramConn) String() (string) {
	return conn.udp.LocalAddr().Network() + ":" + conn.udp.LocalAddr().String()
}

func (conn *EventDatagramConn) getSubscription(event string) (*EventDatagramSub) {
	var subs *EventDatagramSub
	conn.lock.RLock()
	for _, subs = range conn.subs {
		if subs.event == event {
			break
		}
	}
	conn.lock.RUnlock()
	return subs
}


// EventDatagram
type EventDatagram struct {
	mi mi.MI
	lock sync.Mutex
	conns []*EventDatagramConn
}

func (event *EventDatagram) Init(mi mi.MI) (error) {

	/* we typically only use one socket */
	event.conns = make([]*EventDatagramConn, 0, 1)
	event.mi = mi
	return nil
}

func (event *EventDatagram) Subscribe(ev string, fn EventNotification, fnp interface{}) (Subscription) {

	var conn *EventDatagramConn

	/* search for a connection that does not have this event registered */
	event.lock.Lock()
	for _, conn = range event.conns {
		if conn.getSubscription(ev) == nil {
			break
		}
	}

	if conn == nil {
		conn = &EventDatagramConn{}
		conn.Init(event)
		if conn == nil {
			return nil
		}
		/* add the new connection */
		event.conns = append(event.conns, conn)
	}

	sub := &EventDatagramSub{conn: conn, event:ev, fn: fn, fnp: fnp}
	conn.subs = append(conn.subs, sub)
	event.lock.Unlock()

	/* we now have a proper conn to listen for events on */
	logrus.Debug("subscribing for " + sub.Event() + " on " + sub.String())

	/* we've got the connection - let us subscribe */
	var eviParams = map[string]interface{}{
		"event": ev,
		"socket": conn.String(),
		"expire": 120,
	}
	_, err := event.mi.CallSync("event_subscribe", &eviParams);
	if err != nil {
		logrus.Error("could not subscribe for event " + ev)
		sub.Unsubscribe()
		return nil
	}

	logrus.Debug("subscribed " + sub.Event() + " at " + sub.String())
	return sub
}

func (event *EventDatagram) Close() {
	logrus.Debug("closing datagram handler")
}
