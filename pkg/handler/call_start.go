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
	"github.com/google/uuid"
	"github.com/OpenSIPS/opensips-calling-api/pkg/event"
)

type callCoords struct {
	callid, caller, callee string
	ruri, dlginfo string
}

func (h *Handler) callStartEnd(coords *callCoords) {
	var byeParams = map[string]string{
		"method": "BYE",
		"ruri": coords.ruri,
		"headers": coords.dlginfo + "CSeq: 3 BYE\r\n", /* guessing the cseq */
	}
	h.mi.Call("t_uac_dlg", &byeParams, nil, nil)
}

func (h *Handler) callStartNotify(result map[string]interface{}, param interface{}, sub event.Subscription) {
	var ok bool
	var coords *callCoords

	if coords, ok = param.(*callCoords); !ok {
		sub.Unsubscribe()
		h.done <- errors.New("invalid parameter passed at callback")
		return
	}

	status, ok := result["status"].(string)
	if ok != true {
		h.done <- errors.New("invalid returned status type")
		return
	}
	h.Report("call " + coords.callid + " transfering status: " + status);

	switch status[0] {
	case '1': /* provisional - all good */
	case '2': /* transfer successful */
		h.callStartEnd(coords)
		h.done <- nil
	default:
		h.done <- errors.New("Transfer failed with status " + status)
	}
}

func (h *Handler) callStartTransfer(result map[string]interface{}, param interface{}) {

	var ok bool
	var coords *callCoords

	if coords, ok = param.(*callCoords); !ok {
		h.done <- errors.New("invalid parameter passed at callback")
		return
	}

	/* TODO: handle failure */

	/* XXX: report 2 - call transferred */
	h.Report("call " + coords.callid + " has been transfered to " + coords.callee)
}

func (h *Handler) callStartInitial(result map[string]interface{}, param interface{}) {

	var ok bool
	var coords *callCoords

	if coords, ok = param.(*callCoords); !ok {
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

	coords.ruri, ok = result["RURI"].(string)
	if ok != true {
		h.done <- errors.New("invalid RURI returned")
		return
	}

	message, ok := result["Message"].(string)
	if ok != true {
		h.done <- errors.New("invalid Message returned")
		return
	}

	/* gather information about the dialog, so we can close it later */
	for _, header := range strings.Split(message, "\r\n") {
		switch strings.Split(header, ":")[0] {
		case "From", "To", "Routes", "Call-ID", "Call-Id":
			coords.dlginfo += header + "\r\n"
		}
	}

	/* XXX: report 1 - call answered */
	h.Report("call " + coords.callid + " answered by " + coords.caller)

	var transferParams = map[string]string{
		"callid": coords.callid,
		"leg": "callee",
		"destination": coords.callee,
	}

	/* before transfering, register for new blind transfer events */
	subs := h.ev.Subscribe("E_CALL_BLIND_TRANSFER", h.callStartNotify, coords)

	err := h.mi.Call("call_transfer", &transferParams, h.callStartTransfer, coords)
	if err != nil {
		subs.Unsubscribe()
		h.done <- err
		return
	}
}

func (h *Handler) CallStart(params map[string]string) {

	const headersFormat = "From: <%s>\r\n" +
		"To: <%s>\r\n" +
		"Contact: <%s>\r\n" +
		"Content-Type: application/sdp\r\n" +
		"CSeq: 1 INVITE\r\n" +
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
		ruri: caller,
		dlginfo: "",
	}

	err := h.mi.Call("t_uac_dlg", &inviteParams, h.callStartInitial, &coords)
	if err != nil {
		h.done <- err
		return
	}
}
