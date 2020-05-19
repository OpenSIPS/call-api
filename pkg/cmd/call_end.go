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
)

func (c *Cmd) CallEnd(params map[string]string) {

	callid, ok := params["callid"]
	if ok != true {
		c.done <- errors.New("callid not specified")
		return
	}

	var endParams = map[string]string{
		"dialog_id": callid,
	}

	ret, err := c.proxy.MICallSync("dlg_end_dlg", &endParams)
	if err != nil {
		c.done <- err
	} else if ret.IsError() {
		c.done <- ret.Error
	} else {
		c.done <- nil
	}
}
