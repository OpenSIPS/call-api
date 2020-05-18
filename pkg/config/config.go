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

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// OpenSIPS Calling API "config.yml" file structure
type Config struct {
	Server struct {
		Host string `yaml:"host,omitempty"`
		Port int `yaml:"port,omitempty"`
	} `yaml:"server"`

	Log struct {
		FilePath string `yaml:"file_path",omitempty"`
		Level string `yaml:"level",omitempty"`
	} `yaml:"log"`
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
func ParseFlags(tool string) (string, error) {
	var configPath string

	flag.StringVar(&configPath, "config", "./config/" + tool " + .yml", "path to config file")

	flag.Parse()

	return configPath, nil
}


func InitLogging(cfg *Config) (file *os.File, err error) {
	if cfg.Log.FilePath != "" {
		f, err := os.OpenFile(cfg.Log.FilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.SetOutput(f)
		file = f
	}

	if cfg.Log.Level != "" {
		level, err := logrus.ParseLevel(cfg.Log.Level)
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.SetLevel(level)
	}

	return
}
