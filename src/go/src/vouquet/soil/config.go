package soil

import (
	"fmt"
	"path/filepath"
)

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	DBUser   string
	DBPass   string
	DBServer string
	DBName   string
	DBPort   int

	GMO *GMO `toml:GMO`
}

func (self *Config) sqlcred() string {
	cred := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", self.DBUser, self.DBPass,
									self.DBServer, self.DBPort, self.DBName)
	return cred
}

type GMO struct {
	ApiKey string
	SecretKey string
}

func loadConfig(path string) (*Config, error) {
	fpath := filepath.Clean(path)

	var conf Config
	if _, err := toml.DecodeFile(fpath, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
