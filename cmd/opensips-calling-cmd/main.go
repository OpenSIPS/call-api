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
	"flag"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/cmd"
	"github.com/OpenSIPS/opensips-calling-api/pkg/proxy"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/pkg/ws_server"
)

/* used to simulate the Communication interface */
type CmdConnection struct {}

func (conn *CmdConnection) Notify(c *cmd.Cmd, notify interface{}) {
	/* this connection simply outputs the results */
	logrus.Printf("%s: %v", c.ID, notify)
}

func usage(prog string) {
	logrus.Fatalf("Usage: %s command [arguments...]", prog)
}

func main() {

	cfgPath, err := config.ParseFlags("calling-cmd")
	if err != nil {
		logrus.Fatal(err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		logrus.Fatal(err)
	}

	/* initialize logging */
	logfile, err := config.InitLogging(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	if logfile != nil {
		defer logfile.Close()
	}

	if flag.NArg() < 1 {
		logrus.Error("no command specified!")
		usage(os.Args[0])
	}

	proxy := proxy.NewProxy(cfg)
	if proxy == nil {
		logrus.Fatal("could not initialize SIP proxy")
	}
	command := flag.Arg(0)
	logrus.Debugf("Running command %s", command)
	var conn ws_server.Connection = new(CmdConnection)
	c := cmd.New(command, "", proxy, conn.Notify)
	if c == nil {
		logrus.Fatalf("could not initialize %s command", command)
	}
	var arguments = map[string]string{}
	for _, arg := range flag.Args()[1:] {
		param := strings.Split(arg, "=")
		arguments[param[0]] = strings.Join(param[1:], "=")
	}
	c.Run(arguments)
	err = c.Wait()
	if err != nil {
		logrus.Fatal(err.Error())
	}
}
