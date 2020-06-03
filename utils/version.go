//
// Copyright (C) 2020 OpenSIPS Solutions
//
// Call API is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Call API is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
//

package utils

const (
	VERSION = "0.1"
	RELEASE = "beta"
)

var (
	GitCommit string
	BuildTime string
)

func GetFullVersion() (string) {
	var version string

	version = "v" + VERSION
	if RELEASE != "" {
		version += "-" + RELEASE
	}
	if GitCommit != "" {
		version += "-" + GitCommit[0:12]
	}
	if BuildTime != "" {
		version += "@" + BuildTime
	}
	return version
}
