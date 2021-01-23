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

package event

import (
	"errors"
	"net"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/call-api/pkg/mi"
	"github.com/OpenSIPS/call-api/internal/jsonrpc"
)

// DatagramSubscription - object referenced by Event users
type DatagramSubscription struct {
	valid bool
	notify EventNotification
	filter map[string]interface{}
	handler *EventDatagramSub
}

func (sub *DatagramSubscription) Event() (string) {
	return sub.handler.String()
}

func (sub *DatagramSubscription) String() (string) {
	return sub.handler.String() // TODO: add filters
}

func (sub *DatagramSubscription) Unsubscribe() {
	sub.valid = false
	sub.handler.removeSubscription(sub)
}

func (sub *DatagramSubscription) MatchFilter(notify *jsonrpc.JsonRPCNotification) (bool) {
	if sub.filter == nil {
		return true
	}
	for k, v := range sub.filter {
		r, err := notify.Get(k)
		if err != nil || r != v {
			return false
		}
	}
	return true
}

// EventDatagramSub - manages a subscription to an event to the proxy
type EventDatagramSub struct {
	event string
	subscribed bool
	confirm chan error
	lock sync.RWMutex
	handler *EventDatagram
	subscriptions []*DatagramSubscription
}

func (sub *EventDatagramSub) String() (string) {
	return sub.event
}

func (sub *EventDatagramSub) IsSubscribed() (bool) {
	return sub.subscribed
}

func (sub *EventDatagramSub) WaitSubscribed() (bool) {
	if sub.IsSubscribed() {
		return true
	}
	// wait for it to subscribe
	<-sub.confirm
	return sub.IsSubscribed()
}

func (sub *EventDatagramSub) newSubscription(notify EventNotification, filter map[string]interface{}) (*DatagramSubscription) {

	ds := &DatagramSubscription{
		notify: notify,
		filter: filter,
		handler: sub,
		valid: true,
	}
	sub.lock.Lock()
	sub.subscriptions = append(sub.subscriptions, ds)
	sub.lock.Unlock()
	return ds
}

func (sub *EventDatagramSub) removeSubscription(ds *DatagramSubscription) {
	sub.lock.Lock()
	for i, s := range sub.subscriptions {
		if s == ds {
			sub.subscriptions = append(sub.subscriptions[0:i], sub.subscriptions[i+1:]...)
			break
		}
	}
	if len(sub.subscriptions) == 0 {
		sub.handler.removeEventSubscription(sub)
	}
	sub.lock.Unlock()
	if !sub.IsSubscribed() {
		return
	}
	// we now properly unregister
	var eviParams = map[string]interface{}{
		"event": sub.event,
		"socket": sub.handler.String(),
		"expire": 0,
	}
	err := sub.handler.mi.Call("event_subscribe", &eviParams, nil);
	if err != nil {
		logrus.Error("could not unsubscribe for event " + sub.event + " " + err.Error())
	} else {
		logrus.Debug("successfully unsubscribed " + sub.event)
	}
}

func (sub *EventDatagramSub) subscribeReply(response *jsonrpc.JsonRPCResponse) {

	if !response.IsError() {
		// confirm the event is properly subscribed
		sub.subscribed = true
	} else {
		// wake up the event loop to inform there's no one in there
		sub.lock.Lock()
		if len(sub.subscriptions) == 0 {
			sub.handler.removeEventSubscription(sub)
		}
		sub.lock.Unlock()
	}
	close(sub.confirm)
}

func (sub *EventDatagramSub) notify(n *jsonrpc.JsonRPCNotification) {
	sub.lock.RLock()
	for _, s := range sub.subscriptions {
		if s.valid && s.MatchFilter(n) {
			go s.notify(s, n)
		}
	}
	sub.lock.RUnlock()
}

// EventDatagram - handler of the Datagram connection
type EventDatagram struct {
	mi mi.MI
	lock sync.Mutex
	conn *net.UDPConn
	subs []*EventDatagramSub
}

func (event *EventDatagram) waitForEvents() {

	var sub *EventDatagramSub

	buffer := make([]byte, 65535)
	for {
		r, _, err := event.conn.ReadFrom(buffer)
		if err == nil {
			result := &jsonrpc.JsonRPCNotification{}
			err = result.Parse(buffer[0:r])
			if err != nil {
				logrus.Error("could not parse notification: " + err.Error())
			} else {
				sub = event.getEventSubscription(result.Method)
				// run in a different routine to avoid blocking
				if sub != nil {
					sub.notify(result)
				}
			}
		} else {
			logrus.Warn("error while listening for events: " + err.Error())
		}
	}
}

func (event *EventDatagram) getEventSubscription(ev string) (*EventDatagramSub) {
	var es *EventDatagramSub
	for _, es = range event.subs {
		if es.event == ev {
			break
		} else {
			es = nil
		}
	}
	return es
}

func (event *EventDatagram) newEventSubscription(ev string) (*EventDatagramSub) {
	return &EventDatagramSub{
		event: ev,
		handler: event,
		confirm: make(chan error, 1),
		subscriptions: make([]*DatagramSubscription, 0),
	}
}

func (event *EventDatagram) removeEventSubscription(evSub *EventDatagramSub) {
	event.lock.Lock()
	for i, s := range event.subs {
		if evSub == s {
			logrus.Info("removing event " + evSub.String())
			event.subs = append(event.subs[0:i], event.subs[i+1:]...)
			break;
		}
	}
	event.lock.Unlock()
}

func (event *EventDatagram) Init(mi mi.MI) (error) {

	miAddr, ok := mi.Addr().(*net.UDPAddr)
	if ok != true {
		return errors.New("using non-UDP protocol to connect to MI")
	}
	c, err := net.DialUDP("udp", nil, miAddr)
	if err != nil {
		return err
	}
	udpAddr, ok := c.LocalAddr().(*net.UDPAddr)
	if ok != true {
		return errors.New("using non-UDP local socket to connect to MI")
	}
	local := net.UDPAddr{IP: udpAddr.IP}
	udpConn, err := net.ListenUDP(c.LocalAddr().Network(), &local)
	if err != nil {
		return err
	}
	udpConn.SetReadBuffer(65535)
	udpConn.SetWriteBuffer(65535)

	file, _ := udpConn.File()
	fd := file.Fd()
	syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 0)

	event.mi = mi
	event.conn = udpConn
	event.subs = make([]*EventDatagramSub, 0, 1)

	go event.waitForEvents()
	return nil
}

func (event *EventDatagram) SubscribeFilter(ev string, notify EventNotification, filter map[string]interface{}) (Subscription) {

	var newSub bool
	newSub = false

	/* search for a connection that does not have this event registered */
	event.lock.Lock()
	evSub := event.getEventSubscription(ev)
	if evSub == nil {
		evSub = event.newEventSubscription(ev)
		if evSub == nil {
			logrus.Error("could not create new subscription")
			return nil
		} else {
			event.subs = append(event.subs, evSub)
			newSub = true
		}
	}
	event.lock.Unlock()

	if newSub {
		/* we now have a proper conn to listen for events on */
		logrus.Debug("subscribing for " + ev + " on " + event.String())

		/* we've got the connection - let us subscribe */
		var eviParams = map[string]interface{}{
			"event": ev,
			"socket": event.String(),
			"expire": 120,
		}
		err := event.mi.Call("event_subscribe", &eviParams, evSub.subscribeReply)
		if err != nil {
			logrus.Error("could not subscribe for event " + ev + ": " + err.Error())
			event.removeEventSubscription(evSub)
			return nil
		}
	}

	if !evSub.WaitSubscribed() {
		logrus.Error("could not subscribe for event " + ev)
		event.removeEventSubscription(evSub)
		return nil
	}

	return evSub.newSubscription(notify, filter)
}

func (event *EventDatagram) Subscribe(ev string, notify EventNotification) (Subscription) {
	return event.SubscribeFilter(ev, notify, nil)
}

func (event *EventDatagram) String() (string) {
	return event.conn.LocalAddr().Network() + ":" + event.conn.LocalAddr().String()
}
