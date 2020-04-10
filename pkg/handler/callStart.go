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
	"log"
	"errors"
)

func (h *Handler) CallStart(params []string) (string, error) {

	if len(params) != 2 {
		return "", errors.New("caller and/or callee not specified")
	}
	caller := params[0]
	callee := params[1]
	log.Printf("caller=%s callee=%s", caller, callee)

	return "call successfully started", nil
}
