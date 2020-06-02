# JSON-RPC Protocol Documentation

Once a WebSocket channel is established between the client and the API, communication will take place strictly using JSON messages which follow the JSON-RPC 2.0 request/response/notification protocol.  Note that API _clients are expected to process notifications_ from the API, while their launched commands are being handled asynchronously.

The Call API daemon receives a set of commands as JSON-RPC requests. A typical
command would look as it follows (note that `<...>` indicates a placeholder
for the actual values):

```
{
	"jsonrpc": "2.0"
	"id": <request-id>,
	"method": "<command>",
	"params": <params>
}
```

The request will contain the following placeholders:

* `request-id`: an unique identifier of the JSON-RPC request
* `method`: of the commands provided by the Call API engine
* `params`: a JSON object containing different parameters, mandatory or
optional, required by the command to run

For each command sent, the JSON-RPC client will immediately receive a response
from the server, with the following format:

```
{
	"jsonrpc": "2.0"
	"id": <request-id>,
	"result": {
		"cmd_id": "<cmd-id>",
		"event": "Started",
	}
}
```

The JSON above contains the following placeholders:

* `cmd-id`: an identifier of the command sent. If a `cmd_id` node has been
specified in the request, then the same id will be returned otherwise an
unique uuid will be generated for the command; this id would be present in all
the following notifications sent for this specific command.
* `request-id`: is the identifier sent in the JSON-RPC request

In case of a failure to initiate the command, a standard JSON-RPC error will
be triggered, with the following format:

```
{
	"jsonrpc": "2.0"
	"id": <request-id>,
	"error": {
		"code": <code>,
		"message": "<reason>"
	}
}
```

The placeholders for an error will contain the following values:

* `request-id`: the same identifier received in the JSON-RPC request
* `code`: an integer indicated the code of the error
* `reason`: a string containing the reason of the error

If, however, the command invocation was successful, the Call API engine will
start to generate JSON-RPC notifications about the progress of the command.
These notifications look like this:

```
	"jsonrpc": "2.0"
	"method": "<command>",
	"params": {
		"cmd_id": "<cmd-id>",
		"event": "<event>",
		"data": "<data>",
	}
```

A notification will contain the following values as placeholders:

* `command`: the command that triggered this notification
* `cmd-id`: the id of the command, as provided in the initial JSON-RPC
response
* `event`: one of the following values:
  * `Error`: indicates an error has been triggered
  * `Ended`: indicates that the command has been completed
  * an arbitrary name describing the status of the execution
* `data`: optional JSON node, containing extra information about the error, or
the progress of the command being executed; note that the `Ended` event
does not have a `data` node.

# Commands

## CallStart

### Parameters

* _"caller"_ (string, mandatory)
* _"callee"_ (string, mandatory)

### Events

* _CallerAnswered_: triggered when the caller answered the initial call
  * _caller_: the caller that has just answered the call
  * _callee_: the callee that is being reached next
* _Transferring_: triggered when the caller is trying to reach the callee
  * _caller_: the caller of the new call
  * _destination_: the SIP URI that is being called
* _TransferStart_: triggered when the caller starts the call to callee
  * _callid_: the Call-ID of the new call
  * _caller_: the caller of the new call
  * _callee_: the callee that is being called
* _TransferPending_: _optional_, contains extra information provided by the UAC
regarding the new call
  * _callid_: the Call-ID of the new call
  * _caller_: the caller of the new call
  * _callee_: the callee of the new call
  * _extra_: _optional_, extra information provided by the caller
* _CalleeAnswered_: triggered when the callee answers the call
  * _callid_: the Call-ID of the new call
  * _caller_: the caller of the new call
  * _callee_: the callee of the new call

### Example JSON-RPC flow:

```
1) WS client ----------> API

{
    "method": "CallStart",
    "params": {
        "caller": "sip:alice@10.0.0.10",
        "callee": "sip:bob@10.0.0.11"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

2) WS client <---------- API

{
    "result": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Started"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

3) WS client <---------- API

{
    "method": "CallStart",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "CallerAnswered"
        "data": {
            "caller": "sip:alice@10.0.0.10",
            "callee": "sip:bob@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

4) WS client <---------- API

{
    "method": "CallStart",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Calling"
        "data": {
            "caller": "sip:alice@10.0.0.10",
            "callee": "sip:bob@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

5) WS client <---------- API

{
    "method": "CallStart",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "CallStart"
        "data": {
            "callid": "1fc8043a-3f89-49f9-8f8c-4c284faf69e3",
            "caller": "sip:alice@10.0.0.10",
            "callee": "sip:bob@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

6) WS client <---------- API

{
    "method": "CallStart",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "CallAnswered"
        "data": {
            "callid": "1fc8043a-3f89-49f9-8f8c-4c284faf69e3",
            "caller": "sip:alice@10.0.0.10",
            "callee": "sip:bob@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

7) WS client <---------- API

{
    "method": "CallStart",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
        "status": "Ended"
    },
    "jsonrpc": "2.0"
}
```

## CallBlindTransfer

### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the targeted dialog
* _"leg"_ (string, mandatory) - which party to transfer.  Possible values: _"caller"_, _"callee"_
* _"destination"_ (string, mandatory) - SIP URI of the blind transfer target

### Events

* _Transferring_: triggered when the transferrer has accepted the transfer
  * _destination_: the destination URI specified in the request
* _TransferStart_: triggered when the leg starts the call the new destination
  * _callid_: the Call-ID of the new call
  * _destination_: SIP URI of the party that we are transferring to - note
  that the URI might be altered after the lookup has been performed
* _TransferPending_: _optional_, triggered when the participant starts the
transferring call
  * _callid_: the Call-ID of the new call
  * _destination_: SIP URI of the party that we are transferring to
  * _extra_: _optional_, extra information provided by the transferrer
* _TransferSuccessful_: triggered when the destination accepted the new call
  * _callid_: the Call-ID of the new call
  * _destination_: SIP URI of the party that we are transferring to

### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallBlindTransfer",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1",
        "leg": "callee",
        "destination": "sip:cindy@10.0.0.11"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 2) WS client <---------- API

{
    "result": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Started"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 3) WS client <---------- API

{
    "method": "CallBlindTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "Transferring",
        "data": {
            "destination": "sip:cindy@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "CallBlindTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "TransferStart"
        "data": {
            "callid": "29ad6fdb-7806-4c17-b235-02b8f9fad1ae",
            "destination": "sip:cindy@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

# 5) WS client <---------- API

{
    "method": "CallBlindTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "TransferSuccessful"
        "data": {
            "callid": "29ad6fdb-7806-4c17-b235-02b8f9fad1ae",
            "destination": "sip:cindy@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

# 6) WS client <---------- API

{
    "method": "CallBlindTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
        "event": "Ended"
    },
    "jsonrpc": "2.0"
}
```

## CallAttendedTransfer

### Parameters

* _"callidA"_ (string, mandatory) - the SIP Call-ID of the dialog #1
* _"legA"_ (string, mandatory) - which party to transfer from dialog #1.  Possible values: _"caller"_, _"callee"_
* _"callidB"_ (string, mandatory) - the SIP Call-ID of the dialog #2
* _"legB"_ (string, mandatory) - which party to transfer from dialog #2.  Possible values: _"caller"_, _"callee"_

### Events

* _Transferring_: triggered when the transferrer has accepted the transfer
* _TransferStart_: triggered when the transferring leg of Call-ID of the
dialog #1 is calling the leg in dialog #2
  * _callid_: the Call-ID of the new call
  * _destination_: SIP URI of the party that we are transferring to
* _TransferPending_: _optional_, triggered when the participant starts the
transferring call
  * _callid_: the Call-ID of the new call
  * _destination_: SIP URI of the party that we are transferring to
  * _extra_: _optional_, extra information provided by the transferrer
* _TransferSuccessful_: triggered when the destination accepted the new call
  * _callid_: the Call-ID of the new call
  * _destination_: SIP URI of the party that we are transferring to

### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallAttendedTransfer",
    "params": {
        "callidA": "431fc357.a3e3.49c2@127.0.0.1",
        "legA": "caller",
        "callidB": "0ba2cf53-9a78-41de-8fe6-f2e5bb4d1a1e",
        "legB": "callee",
    }

    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 2) WS client <---------- API

{
    "result": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Started"
    }

    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 3) WS client <---------- API

{
    "method": "CallAttendedTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "Transferring"
    },
    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "CallAttendedTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "TransferStart"
        "data": {
            "callid": "29ad6fdb-7806-4c17-b235-02b8f9fad1ae",
            "destination": "sip:cindy@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

# 5) WS client <---------- API

{
    "method": "CallAttendedTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "TransferSuccessful"
        "data": {
            "callid": "29ad6fdb-7806-4c17-b235-02b8f9fad1ae",
            "destination": "sip:cindy@10.0.0.11"
        }
    },
    "jsonrpc": "2.0"
}

# 6) WS client <---------- API

{
    "method": "CallAttendedTransfer",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
        "event": "Ended"
    },
    "jsonrpc": "2.0"
}
```

## CallHold

### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the target dialog

### Events

* _CallHolding_: triggered when the command was received by the proxy
* _CallHoldStart_: triggered when the proxy sends a hold INVITE to one of the
legs
  * _leg_: the call's leg that the message is being sent to (_caller_ or
  _callee_)
* _CallHoldSuccessful_: triggered when one of the legs successfully accepted
the call hold
  * _leg_: the call's leg that completed the call hold (_caller_ or _callee_)

### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallHold",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1",
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 2) WS client <---------- API

{
    "result": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Started"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 3) WS client <---------- API

{
    "method": "CallHold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallHolding"
    },
    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "CallHold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallHoldStart",
        "data": {
            "leg": "caller"
        },
    },
    "jsonrpc": "2.0"
}

# 5) WS client <---------- API

{
    "method": "CallHold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallHoldStart"
        "data": {
            "leg": "callee"
        },
    },
    "jsonrpc": "2.0"
}

# 6) WS client <---------- API

{
    "method": "CallHold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallHoldSuccessful"
        "data": {
            "leg": "caller"
        },
    },
    "jsonrpc": "2.0"
}

# 7) WS client <---------- API

{
    "method": "CallHold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallHoldSuccessful"
        "data": {
            "leg": "callee"
        },
    },
    "jsonrpc": "2.0"
}

# 8) WS client <---------- API

{
    "method": "CallHold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "Ended"
    },
    "jsonrpc": "2.0"
}
```

## CallUnhold

### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the target dialog

### Events

* _CallUnholding_: triggered when the command was received by the proxy
* _CallUnholdStart_: triggered when the proxy sends a hold INVITE to one of the
legs
  * _leg_: the call's leg that the message is being sent to (_caller_ or
  _callee_)
* _CallUnholdSuccessful_: triggered when one of the legs successfully resumes
  * _leg_: the call's leg that resumed the call (_caller_ or _callee_)

### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallUnhold",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1",
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 2) WS client <---------- API

{
    "result": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Started"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 3) WS client <---------- API

{
    "method": "CallUnhold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallUnholding"
    },
    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "CallUnhold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallUnholdStart",
        "data": {
            "leg": "caller"
        },
    },
    "jsonrpc": "2.0"
}

# 5) WS client <---------- API

{
    "method": "CallUnhold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallUnholdStart"
        "data": {
            "leg": "callee"
        },
    },
    "jsonrpc": "2.0"
}

# 6) WS client <---------- API

{
    "method": "CallUnhold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallUnholdSuccessful"
        "data": {
            "leg": "caller"
        },
    },
    "jsonrpc": "2.0"
}

# 7) WS client <---------- API

{
    "method": "CallUnhold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "CallUnholdSuccessful"
        "data": {
            "leg": "callee"
        },
    },
    "jsonrpc": "2.0"
}

# 8) WS client <---------- API

{
    "method": "CallUnhold",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "Ended"
    },
    "jsonrpc": "2.0"
}
```

## CallEnd

### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the target dialog

### Events

*NO events*

### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallEnd",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 2) WS client <---------- API

{
    "result": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Started"
    },
    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

# 3) WS client <---------- API

{
    "method": "CallEnd",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "event": "Ended"
    }
    "jsonrpc": "2.0"
}
```

## Echo

Command that receives arbitrary parameters and outputs them back as a
notification. This command is useful to test connectivity to the API server.

### Parameters

_Any_ parameter

### Events

* _Reply_: contains a JSON object with the same JSON as specified in the request.

### Example JSON-RPC flow:

```
1) WS client ----------> API

{
    "method": "Echo",
    "params": {
        "test": "echo",
    },
    "id": "33f6c98c821b",
    "jsonrpc": "2.0"
}

2) WS client <---------- API

{
    "result": {
        "cmd_id": "0f8a1664-01e2-46fe-ae55-e02715afee02",
        "event": "Started"
    },
    "id": "33f6c98c821b",
    "jsonrpc": "2.0"
}

3) WS client <---------- API

{
    "method": "Echo",
    "params": {
        "cmd_id": "0f8a1664-01e2-46fe-ae55-e02715afee02",
        "event": "Reply",
        "data": {"test":"echo"}
    },
    "jsonrpc": "2.0"
}

4) WS client <---------- API

{
    "method": "Echo",
    "params": {
        "cmd_id": "0f8a1664-01e2-46fe-ae55-e02715afee02",
        "event": "Ended"
    },
    "jsonrpc": "2.0"
}
```
