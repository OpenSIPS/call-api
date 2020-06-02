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

package cmd

import (
	"fmt"
)

type CmdEvent struct {
	Error error
	Name string
	Params interface{}
}

func NewError(err error) (c *CmdEvent) {
	return &CmdEvent{Error: err}
}

func NewEvent(name string, event interface{}) (c *CmdEvent) {
	switch name {
		case
			"End",
			"Started",
			"Error":
			panic("Event '" + name + "' is reserved")
	}
	return &CmdEvent{Name: name, Params: event}
}

func (c *CmdEvent) IsError() (bool) {
	return c.Error != nil
}

func (c *CmdEvent) HasParams() (bool) {
	return c.Params != nil
}

func (c *CmdEvent) String() (string) {
	if c.IsError() {
		return c.Error.Error()
	} else if !c.HasParams() {
		return c.Name
	} else {
		return c.Name + ": " + fmt.Sprint(c.Params)
	}
}
