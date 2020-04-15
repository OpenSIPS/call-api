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
	"github.com/OpenSIPS/opensips-calling-api/pkg/connection"
)

type Handler struct {
	conn *connection.Connection
	mi mi.MI
}

func New(conn *connection.Connection) (h *Handler) {
	h = new(Handler)
	h.mi = mi.MIHandler()
	h.conn = conn
	return
}

func (h *Handler) Run(command string, params map[string]string) {

	f := reflect.ValueOf(h).MethodByName(command)
	if !f.IsValid() {
		(*h.conn).Error(errors.New(command + " not implemented"))
		return
	}
	in := []reflect.Value{reflect.ValueOf(params)}
	ret := f.Call(in)

	if !ret[1].IsNil() {
		var err error = ret[1].Interface().(error)
		(*h.conn).Error(err)
		return
	}
	(*h.conn).Report(ret[0].String())
}

func (h *Handler) Report(report string) {
	(*h.conn).Report(report)
}
