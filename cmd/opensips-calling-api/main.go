//
// Copyright (c) 2020 OpenSIPS Project
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
	"github.com/sirupsen/logrus"
	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/pkg/ws_server"
)


func main() {
	// parse cmdline args
	cfgPath, err := config.ParseFlags("config")
	if err != nil {
		logrus.Fatal(err)
	}

	// read configuration
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		logrus.Fatal(err)
	}

	// prepare logging
	logfile, err := config.InitLogging(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	if logfile != nil {
		defer logfile.Close()
	}

	ws_server.Run(cfg)
}
