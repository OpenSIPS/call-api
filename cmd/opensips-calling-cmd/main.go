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
	"flag"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/pkg/server"
	"github.com/OpenSIPS/opensips-calling-api/pkg/handler"
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

	cfgPath, err := config.ParseFlags("calling-cmd")
	if err != nil {
		logrus.Fatal(err)
	}

	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		logrus.Fatal(err)
	}

	/* initilize logging */
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
	command := flag.Arg(0)
	logrus.Debugf("Running command %s", command)
	var conn server.Connection = new(CmdConnection)
	h := handler.New(cfg, &conn)
	var arguments = map[string]string{}
	for _, arg := range flag.Args()[1:] {
		param := strings.Split(arg, "=")
		arguments[param[0]] = strings.Join(param[1:], "=")
	}
	err = h.Run(command, arguments)
	if err == nil {
		err = h.Wait()
	}
	if err != nil {
		logrus.Printf("ERR: %v", err)
	}
}
