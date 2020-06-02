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
	"github.com/OpenSIPS/call-api/pkg/event"
	"github.com/OpenSIPS/call-api/internal/jsonrpc"
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

	var event string

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

	callid, err := notify.GetString("transfer_callid")
	if err != nil {
		cb.cmd.NotifyError(err)
		return
	}

	switch state {
	case "failure":
		cb.cmd.NotifyNewError("Transfer failed with status " + status)
		return
	case "ok":
		event = "TransferSuccessful"
		status = ""
	case "start":
		event = "TransferStart"
		cb.dst, err = notify.GetString("destination")
		if err != nil {
			cb.cmd.NotifyError(err)
			return
		}
	default:
		event = "TransferPending"
	}

	body :=  map[string]interface{}{
		"callid": callid,
		"destination": cb.dst,
	}

	if status != "" {
		body["extra"] = status
	}
	cb.cmd.NotifyEvent(event, body)

	if state == "ok" {
		cb.callBlindTransferEnd()
		cb.cmd.NotifyEnd()
	}
}

func (cb *callBlindTransferCmd) callBlindTransferReply(response *jsonrpc.JsonRPCResponse) {

	if response.IsError() {
		cb.cmd.NotifyError(response.Error)
		cb.sub.Unsubscribe()
		return
	}

	cb.cmd.NotifyEvent("Transferring", map[string]interface{}{
		"destination": cb.dst,
	})
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

	var transferFilter = map[string]interface{}{
		"callid": callid,
	}

	cb := &callBlindTransferCmd{
		cmd: c,
		callid: callid,
		dst: destination,
	}

	/* before transfering, register for new blind transfer events */
	cb.sub = c.proxy.SubscribeFilter("E_CALL_TRANSFER", cb.callBlindTransferNotify, transferFilter)
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
