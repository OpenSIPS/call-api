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
	"fmt"
	"errors"
	"strings"
)

func (h *Handler) CallStart(params []string) (string, error) {

	const headersFormat = "From: <%s>\r\n" +
		"To: <%s>\r\n" +
		"Contact: <%s>\r\n" +
		"Content-Type: application/sdp\r\n"

	const inviteBody = "v=0\r\n" +
		"o=click-to-dial 0 0 IN IP4 0.0.0.0\r\n" +
		"s=session\r\n" +
		"c=IN IP4 0.0.0.0\r\n" +
		"t=0 0\r\n" +
		"m=audio 9 RTP/AVP 0\r\n" +
		"a=rtpmap:0 PCMU/8000\r\n"

	if len(params) < 2 {
		return "", errors.New("caller and/or callee not specified")
	}
	caller := params[0]
	callee := params[1]

	headers := fmt.Sprintf(headersFormat, caller, callee, caller)

	var inviteParams = map[string]string{
		"method": "INVITE",
		"ruri": caller,
		"headers": headers,
		"body": inviteBody,
	}

	ret, err := h.mi.Call("t_uac_dlg", &inviteParams)
	if err != nil {
		return "", err
	}
	status, stringType := ret["Status"].(string)
	if stringType != true {
		return "", errors.New("invalid returned status type")
	}

	if strings.Split(status, " ")[0] != "200" {
		return "", errors.New(status)
	}

	// all good now, first call has been answered
	log.Printf("status: %v", status)
	return "call successfully started", nil
}
