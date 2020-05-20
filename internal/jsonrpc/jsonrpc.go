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

package jsonrpc

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/sirupsen/logrus"
)

type JsonRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  interface{}            `json:"params,omitempty"`
}

type JsonRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *JsonRPCError          `json:"error,omitempty"`
}

type JsonRPCError struct {
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Data    interface{}             `json:"data,omitempty"`
}

type JsonRPCNotification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  interface{}            `json:"params,omitempty"`
}

func (err *JsonRPCError) Error() (string) {
	return strconv.Itoa(err.Code) + " " + err.Message
}

func NewRequest(id interface{}, method string, params interface{}) (*JsonRPCRequest) {
	if _, ok := id.(uint64); !ok {
		if _, ok := id.(string); !ok {
			logrus.Errorf("unsupported ID type, must be uint64 or string: %s\n", id)
			return nil
		}
	}

	req := &JsonRPCRequest{
		JSONRPC: "2.0",
		ID: id,
		Method: method,
		Params: params,
	}
	return req
}

func NewNotification(method string, params interface{}) (*JsonRPCNotification) {
	notify := &JsonRPCNotification{
		JSONRPC: "2.0",
		Method: method,
		Params: params,
	}
	return notify
}

func (request *JsonRPCRequest) Buffer() ([]byte, error) {
	return json.Marshal(request)
}
func (request *JsonRPCRequest) Parse(bytes []byte) (error) {
	return json.Unmarshal(bytes, request)
}

func (reply *JsonRPCResponse) Parse(bytes []byte) (error) {
	return json.Unmarshal(bytes, reply)
}

func (reply *JsonRPCResponse) IsError() (bool) {
	return reply.Error != nil
}

func getString(i interface{}, name string) (string, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return "", errors.New("result is not a map")
	}
	val, ok := m[name].(string)
	if ok != true {
		return "", errors.New("invalid type for " + name)
	}
	return val, nil
}

func (reply *JsonRPCResponse) GetString(name string) (string, error) {
	return getString(reply.Result, name)
}

func (notify *JsonRPCNotification) GetString(name string) (string, error) {
	return getString(notify.Params, name)
}
