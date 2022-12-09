package config

import tmcfg "github.com/tendermint/tendermint/config"

type Config struct {
	*tmcfg.Config
}

func DefaultConfig() *Config {
	return &Config{
		Config: tmcfg.DefaultConfig(),
	}
}
