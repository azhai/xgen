package models

import (
	"github.com/azhai/gozzo/config"
	"github.com/azhai/xgen/cmd"
	"github.com/azhai/xgen/dialect"
)

var (
	connLoaded = false
	connCfgs   = make(map[string]dialect.ConnConfig)
)

func init() {
	config.PrepareEnv(512)
	SetupConns()
}

func SetupConns() {
	if _, err := cmd.LoadConfigFile(true); err != nil {
		panic(err)
	}
	for _, c := range cmd.GetConnConfigs() {
		connCfgs[c.Key] = c
	}
	connLoaded = true
}

func GetConnConfig(key string) dialect.ConnConfig {
	if connLoaded == false {
		SetupConns()
	}
	if cfg, ok := connCfgs[key]; ok {
		return cfg
	}
	return dialect.ConnConfig{}
}
