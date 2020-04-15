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
)

func (h *Handler) CallEnd(params map[string]string) (string, error) {

	callid, ok := params["callid"]
	if ok != true {
		return "", errors.New("callid not specified")
	}

	var dialogParams = map[string]string{
		"dialog_id": callid,
	}

	_, err := h.mi.Call("dlg_end_dlg", &dialogParams)
	if err != nil {
		return "", err
	}
	return "ok", nil
}
