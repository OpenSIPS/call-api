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

package cmd

import (
)

func (c *Cmd) CallEnd(params map[string]interface{}) {

	callid, ok := params["callid"].(string)
	if !ok {
		c.NotifyNewError("callid not specified")
		return
	}

	var endParams = map[string]string{
		"dialog_id": callid,
	}

	ret, err := c.proxy.MICallSync("dlg_end_dlg", &endParams)
	if err != nil {
		c.NotifyError(err)
	} else if ret.IsError() {
		c.NotifyError(ret.Error)
	} else {
		c.NotifyEnd()
	}
}
