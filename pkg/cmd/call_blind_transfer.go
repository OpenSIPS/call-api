//
// Copyright (C) 2020 OpenSIPS Solutions
//
// OpenSIPS Calling API is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// OpenSIPS Calling API is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
//

package cmd

import (
	"github.com/OpenSIPS/opensips-calling-api/pkg/event"
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
)

type callBlindTransferCmd struct {
	cmd *Cmd
	callid, dst string
	sub event.Subscription
}

func (cb *callBlindTransferCmd) callBlindTransferEnd() {
	var byeParams = map[string]string{
		"dialog_id": cb.callid,
	}
	cb.sub.Unsubscribe()
	cb.cmd.proxy.MICall("dlg_end_dlg", &byeParams, nil)
}

func (cb *callBlindTransferCmd) callBlindTransferNotify(sub event.Subscription, notify *jsonrpc.JsonRPCNotification) {

	state, err := notify.GetString("state")
	if err != nil {
		cb.cmd.NotifyError(err)
		return
	}

	status, err := notify.GetString("status")
	if err != nil {
		cb.cmd.NotifyError(err)
		return
	}
	cb.cmd.NotifyEvent(notify.Params)

	switch state {
	case "ok":
		cb.callBlindTransferEnd()
		cb.cmd.NotifyEnd()
	case "failure":
		cb.cmd.NotifyNewError("Transfer failed with status " + status)
	}
}

func (cb *callBlindTransferCmd) callBlindTransferReply(response *jsonrpc.JsonRPCResponse) {

	if response.IsError() {
		cb.cmd.NotifyError(response.Error)
		cb.sub.Unsubscribe()
		return
	}

	/* XXX: report 2 - call transferred */
	cb.cmd.NotifyEvent("transfered to " + cb.dst)
}

func (c *Cmd) CallBlindTransfer(params map[string]interface{}) {

	callid, ok := params["callid"].(string)
	if !ok {
		c.NotifyNewError("callid not specified")
		return
	}
	leg, ok := params["leg"].(string)
	if !ok {
		c.NotifyNewError("leg not specified")
		return
	}
	destination, ok := params["destination"].(string)
	if !ok {
		c.NotifyNewError("destination not specified")
		return
	}

	var transferParams = map[string]string{
		"callid": callid,
		"leg": leg,
		"destination": destination,
	}

	cb := &callBlindTransferCmd{
		cmd: c,
		callid: callid,
	}

	/* before transfering, register for new blind transfer events */
	cb.sub = c.proxy.Subscribe("E_CALL_TRANSFER", cb.callBlindTransferNotify)
	if cb.sub == nil {
		c.NotifyNewError("Could not subscribe for event")
		return
	}

	err := c.proxy.MICall("call_transfer", &transferParams, cb.callBlindTransferReply)
	if err != nil {
		c.NotifyError(err)
		return
	}
}
