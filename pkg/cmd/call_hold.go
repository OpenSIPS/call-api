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

type callHoldCmd struct {
	cmd *Cmd
	callid string
	sub event.Subscription
}

func (ch *callHoldCmd) callHoldNotify(sub event.Subscription, notify *jsonrpc.JsonRPCNotification) {

	state, err := notify.GetString("state")
	if err != nil {
		ch.cmd.NotifyError(err)
		return
	}

	ch.cmd.NotifyEvent(notify.Params)

	switch state {
	case "ok":
		ch.cmd.NotifyEnd()
	case "failure":
		ch.cmd.NotifyNewError("Transfer failed")
	}
}

func (ch *callHoldCmd) callHoldReply(response *jsonrpc.JsonRPCResponse) {

	if response.IsError() {
		ch.cmd.NotifyError(response.Error)
		ch.sub.Unsubscribe()
		return
	}

	ch.cmd.NotifyEvent(response.Result)
}

func (ch *callHoldCmd) callHoldUnhold(cmd string, params map[string]interface{}) {

	callid, ok := params["callid"].(string)
	if !ok {
		ch.cmd.NotifyNewError("callid not specified")
		return
	}
	ch.callid = callid

	/* before transfering, register for new blind transfer events */
	ch.sub = ch.cmd.proxy.Subscribe("E_CALL_HOLD", ch.callHoldNotify)
	if ch.sub == nil {
		ch.cmd.NotifyNewError("Could not subscribe for event")
		return
	}

	var holdParams = map[string]string{
		"callid": callid,
	}

	err := ch.cmd.proxy.MICall(cmd, &holdParams, ch.callHoldReply)
	if err != nil {
		ch.sub.Unsubscribe()
		ch.cmd.NotifyError(err)
	}
}

func (c *Cmd) CallHold(params map[string]interface{}) {

	ch := &callHoldCmd{
		cmd: c,
	}
	ch.callHoldUnhold("call_hold", params)
}

func (c *Cmd) CallUnhold(params map[string]interface{}) {

	ch := &callHoldCmd{
		cmd: c,
	}
	ch.callHoldUnhold("call_unhold", params)
}
