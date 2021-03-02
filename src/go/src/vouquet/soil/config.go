package soil

import (
	"path/filepath"
)

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	ApiKey string
	SecretKey string
}

func LoadConfig(path string) (*Config, error) {
	fpath := filepath.Clean(path)

	var conf Config
	if _, err := toml.DecodeFile(fpath, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
