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

package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/handler"
	"github.com/OpenSIPS/opensips-calling-api/pkg/server"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
)

/* used to simulate the Communication interface */
type CmdConnection struct {}

func (conn *CmdConnection) Report(report string) {
	/* this connection simply outputs the results */
	logrus.Print(report)
}

func (conn *CmdConnection) Close() {
}

func usage(prog string) {
	logrus.Fatalf("Usage: %s command [arguments...]", prog)
}

func main() {
	/*

	if len(os.Args) < 2 {
		logrus.Print("no command specified!")
		usage(os.Args[0])
	}
	command := os.Args[1]
	var conn server.Connection = new(CmdConnection)
	h := handler.New(&conn)
	var arguments = map[string]string{}
	for _, arg := range os.Args[2:] {
		param := strings.Split(arg, "=")
		arguments[param[0]] = strings.Join(param[1:], "=")
	}
	err := h.Run(command, arguments)
	if err == nil {
		err = h.Wait()
	}
	if err != nil {
		logrus.Printf("ERR: %v", err)
	}
	*/
}
