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
	"github.com/OpenSIPS/call-api/pkg/event"
	"github.com/OpenSIPS/call-api/internal/jsonrpc"
)

type callAttendedTransferCmd struct {
	ended bool
	cmd *Cmd
	callid, dst string
	sub event.Subscription
}

func (ca *callAttendedTransferCmd) callAttendedTransferEnd() {
	var byeParams = map[string]string{
		"dialog_id": ca.callid,
	}
	ca.cmd.proxy.MICall("dlg_end_dlg", &byeParams, nil)
}

func (ca *callAttendedTransferCmd) callAttendedTransferNotify(sub event.Subscription, notify *jsonrpc.JsonRPCNotification) {

	var event string

	state, err := notify.GetString("state")
	if err != nil {
		ca.sub.Unsubscribe()
		ca.cmd.NotifyError(err)
		return
	}
	status, err := notify.GetString("status")
	if err != nil {
		ca.sub.Unsubscribe()
		ca.cmd.NotifyError(err)
		return
	}

	callid, err := notify.GetString("transfer_callid")
	if err != nil {
		ca.cmd.NotifyError(err)
		return
	}

	switch state {
	case "failure":
		ca.sub.Unsubscribe()
		ca.cmd.NotifyNewError("transfer failed with status " + status)
	case "ok":
		if !ca.ended {
			ca.callAttendedTransferEnd()
		}
		event = "TransferSuccessful"
		status = ""
	case "start":
		event = "TransferStart"
		ca.dst, err = notify.GetString("destination")
		if err != nil {
			ca.cmd.NotifyError(err)
			return
		}
	default:
		event = "TransferPending"
		// this is a provisional that has a sip status in it - if 200, then we
		// should terminate the initial dialog
		if status != "" && status[0] == '2' {
			ca.callAttendedTransferEnd()
			ca.ended = true
		}
	}

	body :=  map[string]interface{}{
		"callid": callid,
	}
	if ca.dst != "" {
		body["destination"] = ca.dst
	}
	if status != "" {
		body["extra"] = status
	}
	ca.cmd.NotifyEvent(event, body)
	if state == "ok" {
		ca.cmd.NotifyEnd()
	}
}

func (ca *callAttendedTransferCmd) callAttendedTransferReply(response *jsonrpc.JsonRPCResponse) {

	if response.IsError() {
		ca.cmd.NotifyError(response.Error)
		ca.sub.Unsubscribe()
		return
	}

	ca.cmd.NotifyEvent("Transferring", nil)
}

func (c *Cmd) CallAttendedTransfer(params map[string]interface{}) {

	callidA, ok := params["callidA"].(string)
	if !ok {
		c.NotifyNewError("callidA not specified")
		return
	}
	legA, ok := params["legA"].(string)
	if !ok {
		c.NotifyNewError("legA not specified")
		return
	}
	callidB, ok := params["callidB"].(string)
	if !ok {
		c.NotifyNewError("callidB not specified")
		return
	}
	legB, ok := params["legB"].(string)
	if !ok {
		c.NotifyNewError("legB not specified")
		return
	}

	var transferParams = map[string]string{
		"callid": callidA,
		"leg": legA,
		"transfer_callid": callidB,
		"transfer_leg": legB,
	}

	var transferFilter = map[string]interface{}{
		"callid": callidA,
	}

	ca := &callAttendedTransferCmd{
		cmd: c,
		callid: callidA,
	}

	/* before transfering, register for new blind transfer events */
	ca.sub = c.proxy.SubscribeFilter("E_CALL_TRANSFER", ca.callAttendedTransferNotify, transferFilter)
	if ca.sub == nil {
		c.NotifyNewError("Could not subscribe for event")
		return
	}

	err := c.proxy.MICall("call_transfer", &transferParams, ca.callAttendedTransferReply)
	if err != nil {
		ca.sub.Unsubscribe()
		c.NotifyError(err)
		return
	}
}
