
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
	"fmt"
	"strings"
	"time"
	"github.com/OpenSIPS/call-api/pkg/event"
	"github.com/OpenSIPS/call-api/internal/jsonrpc"
)

type callStartCmd struct {
	caller, callee, ruri, dlginfo string
	sub event.Subscription
	cmd *Cmd
}

func (cs *callStartCmd) callStartEnd() {
	var byeParams = map[string]string{
		"method": "BYE",
		"ruri": cs.ruri,
		"headers": cs.dlginfo + "CSeq: 3 BYE\r\n", /* guessing the cseq */
	}
	cs.sub.Unsubscribe()
	cs.cmd.proxy.MICall("t_uac_dlg", &byeParams, nil)
}


func (cs *callStartCmd) callStartNotify(sub event.Subscription, notify *jsonrpc.JsonRPCNotification) {

	var event string

	state, err := notify.GetString("state")
	if err != nil {
		cs.cmd.NotifyError(err)
		return
	}

	status, err := notify.GetString("status")
	if err != nil {
		cs.cmd.NotifyError(err)
		return
	}

	callid, err := notify.GetString("transfer_callid")
	if err != nil {
		cs.cmd.NotifyError(err)
		return
	}

	switch state {
	case "failure":
		cs.cmd.NotifyNewError("transfer failed with status " + status)
		return
	case "ok":
		event = "CalleeAnswered"
		status = ""
	case "start":
		event = "TransferStart"
	default:
		event = "TransferPending"
	}

	body :=  map[string]interface{}{
		"callid": callid,
		"caller": cs.caller,
		"callee": cs.callee,
	}

	if status != "" {
		body["extra"] = status
	}
	cs.cmd.NotifyEvent(event, body)

	if state == "ok" {
		cs.callStartEnd()
		cs.cmd.NotifyEnd()
	}
}

func (cs *callStartCmd) callStartTransfer(response *jsonrpc.JsonRPCResponse) {

	if response.IsError() {
		cs.callStartEnd()
		cs.cmd.NotifyError(response.Error)
		return
	}

	cs.cmd.NotifyEvent("Transferring", map[string]interface{}{
		"caller": cs.caller,
		"destination": cs.callee,
	})
}


func (cs *callStartCmd) callStartInitial(response *jsonrpc.JsonRPCResponse) {

	if response.IsError() {
		cs.cmd.NotifyError(response.Error)
		return
	}

	status, err := response.GetString("Status")
	if err != nil {
		cs.cmd.NotifyError(err)
		return
	}

	if strings.Split(status, " ")[0] != "200" {
		cs.cmd.NotifyNewError("failed to establish initial call: " + status)
		return
	}

	cs.ruri, err = response.GetString("RURI")
	if err != nil {
		cs.cmd.NotifyError(err)
		return
	}

	message, err := response.GetString("Message");
	if err != nil {
		cs.cmd.NotifyError(err)
		return
	}

	/* gather information about the dialog, so we can close it later */
	for _, header := range strings.Split(message, "\r\n") {
		switch strings.Split(header, ":")[0] {
		case "From", "To", "Routes", "Call-ID", "Call-Id":
			cs.dlginfo += header + "\r\n"
		}
	}

	cs.cmd.NotifyEvent("CallerAnswered", map[string]interface{}{
		"caller": cs.caller,
		"callee": cs.callee,
	})

	var transferParams = map[string]string{
		"callid": cs.cmd.ID,
		"leg": "callee",
		"destination": cs.callee,
	}

	var transferFilter = map[string]interface{}{
		"callid": cs.cmd.ID,
	}

	/* before transfering, register for new blind transfer events */
	cs.sub = cs.cmd.proxy.SubscribeFilter("E_CALL_TRANSFER", cs.callStartNotify, transferFilter)
	if cs.sub == nil {
		cs.cmd.NotifyNewError("Could not subscribe for event")
		return
	}

	time.Sleep(500 * time.Millisecond)
	err = cs.cmd.proxy.MICall("call_transfer", &transferParams, cs.callStartTransfer)
	if err != nil {
		cs.sub.Unsubscribe()
		cs.cmd.NotifyError(err)
		return
	}
}

func (c *Cmd) CallStart(params map[string]interface{}) {

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

	caller, ok := params["caller"].(string)
	if !ok {
		c.NotifyNewError("caller not specified")
		return
	}
	callee, ok := params["callee"].(string)
	if !ok {
		c.NotifyNewError("callee not specified")
		return
	}

	headers := fmt.Sprintf(headersFormat, caller, callee, caller, c.ID)

	var inviteParams = map[string]string{
		"method": "INVITE",
		"ruri": caller,
		"headers": headers,
		"body": inviteBody,
	}
	var next_hop = c.proxy.GetURI()
	if next_hop != "" {
		inviteParams["next_hop"] = next_hop
	}

	cs := &callStartCmd{
		caller: caller,
		callee: callee,
		ruri: caller,
		dlginfo: "",
		cmd: c,
	}

	err := c.proxy.MICall("t_uac_dlg", &inviteParams, cs.callStartInitial)
	if err != nil {
		c.NotifyError(err)
		return
	}
}
