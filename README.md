# Call API

----

The Call API is a front-end for SIP Proxies (such as [OpenSIPS](https://opensips.org)), aiming to simplify the management of more advanced SIP call flows.  Combining built-in SIP scenarios (such as the ones from [RFC 5359](https://tools.ietf.org/html/rfc5359)) with real-time notifications as the call commands take place, the API is meant to help VoIP system developers build complex SIP services with ease, altogether while providing live reporting for such services.

The API listens for [WebSocket](https://en.wikipedia.org/wiki/WebSocket) connections on `ws://localhost:5059/call-api` and talks [JSON-RPC 2.0](https://www.jsonrpc.org/specification) over them.

## Requirements

The Call API tool is using go modules, introduced in go 1.13, but the tool was developed based on go version 1.14. We recommend you install at least go 1.14 using your distribution's repositories, or from the official [Golang repository](https://golang.org/dl/).

## Installation

```
    go get github.com/OpenSIPS/call-api
    cd ${GOPATH:-$HOME/go}/src/github.com/OpenSIPS/call-api
    make build
    bin/call-api
```

The following steps will build all the project's tools in the `bin/` folder of the current directory.
 In order to make a complete install of the project, you can follow these steps:

```
    make install
    export PATH=$PATH:${GOBIN:-${GOPATH:-$HOME/go}/bin}
    call-api
```

## Tools

The project builds and installs the following tools:

* **[call-api](cmd/call-api/main.go)** - a WebSocket Server that listens for JSON-RPC requests
* **[call-cmd](cmd/call-cmd/main.go)** - a command line tool that runs the command without  having the `call-api` server running
* **[call-api-client](cmd/call-api-client/main.go)** - a command line tool that sends JSON-RPC requests to the `call-api` daemon and prints the notifications at stdout

## Configuration

Each tool uses a configuration file that defaults to `tool-name.yml` (ex: `call-api.yml` or `call-cmd.yml`). This file is automatically searched in the following folders: `config/`, `/etc` and `/etc/call-api`, in this specific order. A custom path can be specified using the `-config cfg_file.yml` parameter when starting the tool (ex: `call-api -config /etc/custom-config.yml`

Examples of configuration files can be found in the [config](config/) directory.

## API Call Commands

Below are the API's [commands](docs/Commands.md) available for building your JSON-RPC requests.  Read the documentation of each command for a listing of its input parameters and their accepted values:

* **[CallStart](docs/Commands.md#callstart)** - start a call between two participants
* **[CallBlindTransfer](docs/Commands.md#callblindtransfer)** - perform an unattended call transfer (see [RFC 5359 example](https://tools.ietf.org/html/rfc5359#page-50))
* **[CallAttendedTransfer](docs/Commands.md#callattendedtransfer)** - perform an attended call transfer (see [RFC 5359 example](https://tools.ietf.org/html/rfc5359#page-58))
* **[CallHold](docs/Commands.md#callhold)** - put one or both participants on hold
* **[CallUnHold](docs/Commands.md#callunhold)** - resume an on-hold call
* **[CallEnd](docs/Commands.md#callend)** - terminate an ongoing call

## Interacting with the API

By default, the API listens on `ws://localhost:5059/call-api`.  Below is an example way of launching a `CallStart` command using an [API client written in Go](cmd/call-api-client/main.go):

```
go run cmd/call-api-client/main.go \
  -method CallStart \
  -params '{"caller": "sip:alice@localhost", "callee": "sip:bob@localhost"}'
```

The same behavior can be done using the [Call cmd](cmd/call-cmd/main.go) tool:
```
go run cmd/call-cmd/main.go CallStart caller=sip:alice@localhost callee=sip:bob@localhost
```

## Documentation

The [docs](docs/) folder contains the documentation for this project.
