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
	"github.com/OpenSIPS/opensips-calling-api/pkg/event"
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
)

type callAttendedTransferCmd struct {
	ended bool
	legs int
	cmd *Cmd
	callid string
	sub event.Subscription
}

func (ca *callAttendedTransferCmd) callAttendedTransferEnd() {
	var byeParams = map[string]string{
		"dialog_id": ca.callid,
	}
	ca.cmd.proxy.MICall("dlg_end_dlg", &byeParams, nil)
}

func (ca *callAttendedTransferCmd) callAttendedTransferNotify(sub event.Subscription, notify *jsonrpc.JsonRPCNotification) {

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
	ca.cmd.NotifyEvent(notify.Params)

	switch state {
	case "ok":
		ca.legs -= 1
		if ca.legs == 0 {
			if !ca.ended {
				ca.callAttendedTransferEnd()
			}
			ca.cmd.NotifyEnd()
		}
	case "failure":
		ca.legs -= 1
		if ca.legs == 0 {
			ca.sub.Unsubscribe()
			ca.cmd.NotifyNewError("transfer failed with status " + status)
		}
	case "start":
		/* we are counting the legs we will be notified for */
		ca.legs += 1
	default:
		// this is a provisional that has a sip status in it - if 200, then we
		// should terminate the initial dialog
		if status != "" && status[0] == '2' {
			ca.callAttendedTransferEnd()
			ca.ended = true
		}
	}
}

func (ca *callAttendedTransferCmd) callAttendedTransferReply(response *jsonrpc.JsonRPCResponse) {

	if response.IsError() {
		ca.cmd.NotifyError(response.Error)
		ca.sub.Unsubscribe()
		return
	}

	/* XXX: report 2 - call transferred */
	ca.cmd.NotifyEvent("transfering")
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

	ca := &callAttendedTransferCmd{
		cmd: c,
		callid: callidA,
	}

	/* before transfering, register for new blind transfer events */
	ca.sub = c.proxy.Subscribe("E_CALL_TRANSFER", ca.callAttendedTransferNotify)
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
