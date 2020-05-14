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
	"log"

	"github.com/OpenSIPS/opensips-calling-api/pkg/config"
	"github.com/OpenSIPS/opensips-calling-api/pkg/server"
)

func main() {
	// parse cmdline args
	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	// read configuration
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	(&server.CallingAPI{}).Run(cfg)
}
