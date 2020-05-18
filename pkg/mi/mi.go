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

package mi

import (
	"net"
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/internal/jsonrpc"
)

type MIreply func(response *jsonrpc.JsonRPCResponse, param interface{})

var default_url string = "127.0.0.1:8080"

type MI interface {
	Addr() (net.Addr)
	Connect(url string) (error)
	Wait() (error)
	Call(command string, params interface{}, fn MIreply, fnp interface{}) (error)
	CallSync(command string, params interface{}) (*jsonrpc.JsonRPCResponse, error)
}

func MIHandler(config *config.Config) (MI) {
	/* TODO: make a wiser detection here when/if we have multiple backends */
	var url string

	mi := new(MIDatagram)
	if config.MI.URL != "" {
		url = config.MI.URL
	} else {
		url = default_url
		logrus.Debugf("using default url=%s", url)
	}
	if err := mi.Connect(url); err != nil {
		logrus.Printf("ERROR connecting: %v", err)
		return nil
	}
	return mi
}
