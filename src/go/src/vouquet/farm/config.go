package farm

import (
	"fmt"
	"path/filepath"
)

import (
	"github.com/BurntSushi/toml"
)

import (
	"vouquet/shop"
)

type Config struct {
	shop.ConfBase

	DBUser   string
	DBPass   string
	DBServer string
	DBName   string
	DBPort   int

	Works  []*Work `toml:"Work"`
}

func (self *Config) sqlcred() string {
	cred := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", self.DBUser, self.DBPass,
									self.DBServer, self.DBPort, self.DBName)
	return cred
}

func LoadConfig(path string) (*Config, error) {
	fpath := filepath.Clean(path)

	var conf Config
	if _, err := toml.DecodeFile(fpath, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

type Work struct {
	Florist  string  `toml:florist`
	Soil     string  `toml:soil`
	Seed     string  `toml:seed`
	Size     float64 `toml:size`
}
