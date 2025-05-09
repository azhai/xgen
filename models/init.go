package models

import (
	"github.com/azhai/gozzo/config"
	"github.com/azhai/xgen/cmd"
	"github.com/azhai/xgen/dialect"
)

var connCfgs = make(map[string]dialect.ConnConfig)

func PrepareConns(root *config.RootConfig) {
	settings, err := cmd.GetDbSettings(root)
	if err != nil {
		panic(err)
	}
	for _, c := range settings.GetConns() {
		connCfgs[c.Key] = c
	}
}

func GetConnConfig(key string) dialect.ConnConfig {
	if cfg, ok := connCfgs[key]; ok {
		return cfg
	}
	return dialect.ConnConfig{}
}
