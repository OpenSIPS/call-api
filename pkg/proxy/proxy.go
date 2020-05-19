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

package proxy

import (
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/pkg/event"
	"github.com/OpenSIPS/opensips-calling-api/pkg/mi"
)

type Proxy struct {
	mi mi.MI
	ev event.Event
	cfg *config.Config
}

func NewProxy(cfg *config.Config) (p *Proxy) {
	p = &Proxy{cfg: cfg}
	p.mi = mi.MIHandler(cfg)
	if p.mi == nil {
		logrus.Error("could not create MI handler")
		return nil
	}
	p.ev = event.EventHandler(p.mi)
	if p.ev == nil {
		logrus.Error("could not create event handler")
		return nil
	}
	return p
}

func (proxy *Proxy) MICall(command string, params interface{}, fn mi.MIreply) (error) {
	return proxy.mi.Call(command, params, fn)
}

func (proxy *Proxy) MICallSync(command string, params interface{}) (*jsonrpc.JsonRPCResponse, error) {
	return proxy.mi.CallSync(command, params)
}

func (proxy *Proxy) Subscribe(event string, notify event.EventNotification) (event.Subscription) {
	return proxy.ev.Subscribe(event, notify)
}
