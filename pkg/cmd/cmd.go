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

package cmd

import (
	"reflect"
	"github.com/google/uuid"
	"github.com/OpenSIPS/opensips-calling-api/pkg/proxy"
)

type Notify func(cmd *Cmd, notify interface{})


type Cmd struct {
	ID string
	Command string

	proxy *proxy.Proxy
	notify Notify
	done chan error
	hdl reflect.Value
}

func New(command string, id string, p *proxy.Proxy, notify Notify) (c *Cmd) {
	c = &Cmd{
		Command: command,
		ID: id,
		proxy: p,
		notify: notify,
		done: make(chan error, 1),
	}

	if c.ID == "" {
		c.ID = uuid.New().String()
	}

	c.hdl = reflect.ValueOf(c).MethodByName(command)
	if !c.hdl.IsValid() {
		return nil
	}
	return c
}

func (c *Cmd) Run(params map[string]string) {
	in := []reflect.Value{reflect.ValueOf(params)}
	go c.hdl.Call(in)
}

func (c *Cmd) Wait() (error) {

	return <- c.done
}

func (c *Cmd) RunSync(params map[string]string) (error) {

	c.Run(params)
	return c.Wait()
}

func (c *Cmd) Notify(notify interface{}) {
	c.notify(c, notify)
}
