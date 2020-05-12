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
	"fmt"
	"errors"
	"strings"
//	"strconv"
	"github.com/google/uuid"
)

type callCoords struct {
	callid, caller, callee string
}

func (h *Handler) callTransfer(result map[string]interface{}, param interface{}) {

	var ok bool
	var coords callCoords

	if coords, ok = param.(callCoords); !ok {
		h.done <- errors.New("invalid parameter passed at callback")
		return
	}
	h.Report("call " + coords.callid + " has been transfered to " + coords.callee)

	h.done <- nil
}

func (h *Handler) initialCallStart(result map[string]interface{}, param interface{}) {

	var ok bool
	var coords callCoords

	if coords, ok = param.(callCoords); !ok {
		h.done <- errors.New("invalid parameter passed at callback")
		return
	}

	status, ok := result["Status"].(string)
	if ok != true {
		h.done <- errors.New("invalid returned status type")
		return
	}

	if strings.Split(status, " ")[0] != "200" {
		h.done <- errors.New("failed to establish initial call: " + status)
		return
	}

	h.Report("call " + coords.callid + " answered by " + coords.caller)

	var transferParams = map[string]string{
		"callid": coords.callid,
		"leg": "callee",
		"destination": coords.callee,
	}

	err := h.mi.Call("call_transfer", &transferParams, h.callTransfer, coords)
	if err != nil {
		h.done <- err
		return
	}
}

func (h *Handler) CallStart(params map[string]string) {

	const headersFormat = "From: <%s>\r\n" +
		"To: <%s>\r\n" +
		"Contact: <%s>\r\n" +
		"Content-Type: application/sdp\r\n" +
		"Call-Id: %s\r\n"

	const inviteBody = "v=0\r\n" +
		"o=click-to-dial 0 0 IN IP4 0.0.0.0\r\n" +
		"s=session\r\n" +
		"c=IN IP4 0.0.0.0\r\n" +
		"t=0 0\r\n" +
		"m=audio 9 RTP/AVP 0\r\n" +
		"a=rtpmap:0 PCMU/8000\r\n"

	callid := uuid.New().String()

	caller, ok := params["caller"]
	if ok != true {
		h.done <- errors.New("caller not specified")
		return
	}
	callee, ok := params["callee"]
	if ok != true {
		h.done <- errors.New("callee not specified")
		return
	}

	headers := fmt.Sprintf(headersFormat, caller, callee, caller, callid)

	var inviteParams = map[string]string{
		"method": "INVITE",
		"ruri": caller,
		"headers": headers,
		"body": inviteBody,
	}

	coords := callCoords{
		callid: callid,
		caller: caller,
		callee: callee,
	}

	err := h.mi.Call("t_uac_dlg", &inviteParams, h.initialCallStart, coords)
	if err != nil {
		h.done <- err
		return
	}
}
