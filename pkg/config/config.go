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

package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// OpenSIPS Calling API "config.yml" file structure
// Add a line for each YAML setting you want to be decoded
//
// More complex YAML structure examples here:
//   https://github.com/koddr/example-go-config-yaml/blob/master/main.go#L18
type Config struct {
	Host string `yaml:"host,omitempty"`
	Port int `yaml:"port,omitempty"`
}


// read & parse configuration file
func NewConfig(configPath string) (*Config, error) {
	if err := ValidateConfigPath(configPath); err != nil {
		return nil, err
	}

	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}


func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}


// OpenSIPS Calling API command-line parameters
func ParseFlags() (string, error) {
	var configPath string

	flag.StringVar(&configPath, "config", "./config/config.yml", "path to config file")

	flag.Parse()

	return configPath, nil
}
