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
	"errors"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/OpenSIPS/opensips-calling-api/pkg/proxy"
)

type Notify func(cmd *Cmd, notify interface{})


type Cmd struct {
	ID string
	Command string

	proxy *proxy.Proxy
	notify chan *CmdEvent
	hdl reflect.Value
}

func New(command string, id string, p *proxy.Proxy) (c *Cmd) {
	c = &Cmd{
		Command: command,
		ID: id,
		proxy: p,
		notify: make(chan *CmdEvent, 1),
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

func (c *Cmd) Run(params map[string]interface{}) (err error) {
	// TODO: remove this check once non-strings are handled under the hood
	for key := range params {
		if _, ok := params[key].(string); !ok {
			err = fmt.Errorf("non-string parameter values are not yet supported")
			return
		}
	}

	in := []reflect.Value{reflect.ValueOf(params)}
	go c.hdl.Call(in)
	return
}

func (c *Cmd) Wait() (chan *CmdEvent) {

	return c.notify
}

func (c *Cmd) RunSync(params map[string]interface{}) (error) {

	c.Run(params)
	for {
		event := <-c.Wait()
		if event == nil {
			return nil
		} else if event.IsError() {
			return event.Error
		}
	}
}

/* Notify an arbitrary event */
func (c *Cmd) Notify(ce *CmdEvent) {
	c.notify <- ce
}

/* Notify an existing error - closes the channel */
func (c *Cmd) NotifyError(err error) {
	c.Notify(NewError(err))
	close(c.notify)
}

/* Notify a new error - closes the channel */
func (c *Cmd) NotifyNewError(err string ) {
	c.Notify(NewError(errors.New(err)))
	close(c.notify)
}

/* Notify an event */
func (c *Cmd) NotifyEvent(event interface{}) {
	c.Notify(NewEvent(event))
}

/* Notify the termination of the command handling */
func (c *Cmd) NotifyEnd() {
	close(c.notify)
}
