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

package handler

import (
	"errors"
	"reflect"
	"github.com/OpenSIPS/opensips-calling-api/pkg/mi"
	"github.com/OpenSIPS/opensips-calling-api/pkg/event"
	"github.com/OpenSIPS/opensips-calling-api/pkg/connection"
)

type Handler struct {
	conn *connection.Connection
	mi mi.MI
	ev event.Event
	done chan error
}

func New(conn *connection.Connection) (h *Handler) {
	h = new(Handler)
	h.mi = mi.MIHandler()
	h.ev = event.EventHandler(h.mi)
	h.conn = conn
	h.done = make(chan error, 1)
	return h
}

func (h *Handler) Run(command string, params map[string]string) (error) {
	f := reflect.ValueOf(h).MethodByName(command)
	if !f.IsValid() {
		return errors.New(command + " not implemented")
	}
	in := []reflect.Value{reflect.ValueOf(params)}
	go f.Call(in)
	return nil
}

func (h *Handler) Wait() (error) {

	return <- h.done
}

func (h *Handler) RunSync(command string, params map[string]string) (error) {

	err := h.Run(command, params)
	if err != nil {
		return err
	}
	return h.Wait()
}

func (h *Handler) Report(report string) {
	(*h.conn).Report(report)
}
