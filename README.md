# Call API

----

The Call API is a front-end layer for SIP Proxies (such as [OpenSIPS](www.opensips.org)) managing more advanced SIP call flows. Combining built-in SIP scenarios (such as the ones from [RFC 5359](https://tools.ietf.org/html/rfc5359)) with real-time notifications as the call commands take place, the API is meant to help VoIP system developers build complex SIP services with ease, altogether while providing live reporting for such services.

The API listens for [WebSocket](https://en.wikipedia.org/wiki/WebSocket) connections on `ws://localhost:5059/call-api` and talks [JSON-RPC 2.0](https://www.jsonrpc.org/specification) over them.

## Installation

```
    go get github.com/OpenSIPS/call-api
```
**TODO**

## API Call Commands

Below are the API's commands available for building your JSON-RPC requests.  Read the documentation of each command for a listing of its input parameters and their accepted values:

* **[CallStart](#callstart)** - start a call between two participants
* **[CallBlindTransfer](#callblindtransfer)** - perform an unattended call transfer (see [RFC 5359 example](https://tools.ietf.org/html/rfc5359#page-50))
* **[CallAttendedTransfer](#callattendedtransfer)** - perform an attended call transfer (see [RFC 5359 example](https://tools.ietf.org/html/rfc5359#page-58))
* **[CallHold](#callhold)** - put one or both participants on hold
* **[CallUnHold](#callunhold)** - resume an on-hold call
* **[CallEnd](#callend)** - terminate an ongoing call

## Interacting with the API

By default, the API listens on `ws://localhost:5059/call-api`.  Below is an example way of launching a `CallStart` command using an [API client written in Go](cmd/call-api-client/main.go):

```
go run cmd/call-api-client/main.go \
  -method CallStart \
  -params '{"caller": "sip:alice@localhost", "callee": "sip:bob@localhost"}'
```

## JSON-RPC Method Documentation

Once a WebSocket channel is established between the client and the API, communication will take place strictly using JSON messages which follow the JSON-RPC 2.0 request/response/notification protocol.  Note that API _clients are expected to process notifications_ from the API, while their launched commands are being handled asynchronously.

### CallStart

#### Parameters

* _"caller"_ (string, mandatory)
* _"callee"_ (string, mandatory)

#### Example JSON-RPC flow:

```
1) WS client ----------> API

{
    "method": "CallStart",
    "params": {
        "caller": "sip:alice@10.0.0.10",
        "callee": "sip:bob@10.0.0.11"
    }

    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

2) WS client <---------- API

{
    "result": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "status": "Started"
    }

    "id": "831717ed97e5",
    "jsonrpc": "2.0"
}

3) WS client <---------- API

{
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "CALL_ANSWER"
    }

    "jsonrpc": "2.0"
}

4) WS client <---------- API

{
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "TRANSFER_ACCEPT"
    }

    "jsonrpc": "2.0"
}

5) WS client <---------- API

{
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "CALL_START"
    }

    "jsonrpc": "2.0"
}

6) WS client <---------- API

{
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "CALL_ANSWER"
    }

    "jsonrpc": "2.0"
}

7) WS client <---------- API

{
    "method": "Ended",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
    }

    "jsonrpc": "2.0"
}
```

### CallBlindTransfer

#### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the targeted dialog
* _"leg"_ (string, mandatory) - which party to transfer.  Possible values: _"caller"_, _"callee"_
* _"destination"_ (string, mandatory) - SIP URI of the blind transfer target

#### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallBlindTransfer",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1",
        "leg": "callee",
        "destination": "sip:cindy@10.0.0.11"
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
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "TRANSFER_ACCEPT"
    }

    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "CALL_START"
    }

    "jsonrpc": "2.0"
}

# 5) WS client <---------- API

{
    "method": "Ended",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
    }

    "jsonrpc": "2.0"
}
```

### CallAttendedTransfer

#### Parameters

* _"callid_a"_ (string, mandatory) - the SIP Call-ID of the dialog #1
* _"leg_a"_ (string, mandatory) - which party to transfer from dialog #1.  Possible values: _"caller"_, _"callee"_
* _"callid_b"_ (string, mandatory) - the SIP Call-ID of the dialog #2
* _"leg_b"_ (string, mandatory) - which party to transfer from dialog #2.  Possible values: _"caller"_, _"callee"_
* _"destination"_ (string, mandatory) - SIP URI of the attended transfer's destination

#### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallAttendedTransfer",
    "params": {
        "callid_a": "431fc357.a3e3.49c2@127.0.0.1",
        "leg_a": "caller",
        "callid_b": "0ba2cf53-9a78-41de-8fe6-f2e5bb4d1a1e",
        "leg_b": "callee",
        "destination": "sip:cindy@10.0.0.11"
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
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "TRANSFER_ACCEPT"
    }

    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "CALL_START"
    }

    "jsonrpc": "2.0"
}

# 5) WS client <---------- API

{
    "method": "Ended",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
    }

    "jsonrpc": "2.0"
}
```

### CallHold

#### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the target dialog
* _"leg"_ (string, optional) - party to put on hold (_"caller"_ or _"callee"_).  If missing, both call participants will be put on hold

#### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallHold",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1",
        "leg": "caller"
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
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "CALL_HOLD"
    }

    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "Ended",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
    }

    "jsonrpc": "2.0"
}
```

### CallUnHold

#### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the target dialog

#### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallUnHold",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1"
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
    "method": "Event",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178",
        "data": "CALL_UNHOLD"
    }

    "jsonrpc": "2.0"
}

# 4) WS client <---------- API

{
    "method": "Ended",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
    }

    "jsonrpc": "2.0"
}
```

### CallEnd

#### Parameters

* _"callid"_ (string, mandatory) - the SIP Call-ID of the target dialog

#### Example JSON-RPC flow:

```
# 1) WS client ----------> API

{
    "method": "CallEnd",
    "params": {
        "callid": "431fc357.a3e3.49c2@127.0.0.1"
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
    "method": "Ended",
    "params": {
        "cmd_id": "b8179f1e-b4e4-4ac7-9990-4bf64f084178"
    }

    "jsonrpc": "2.0"
}
```
