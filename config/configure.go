package config

import (
	"fmt"

	reverse "github.com/azhai/xgen"
	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/utils"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

var (
	settings *RootConfig
	logger   *utils.Logger
)

// RootConfig 顶层配置，包含其他配置块
type RootConfig struct {
	Debug   bool                   `hcl:"debug,optional" json:"debug,omitempty"`
	App     AppConfig              `hcl:"app,block" json:"app"`
	Reverse reverse.ReverseConfig  `hcl:"reverse,block" json:"reverse,omitempty"`
	Repeats []dialect.RepeatConfig `hcl:"repeat,block" json:"repeat"`
	Conns   []dialect.ConnConfig   `hcl:"conn,block" json:"conn"`
}

// AppConfig App配置块，包括App名称、默认日志和自定义配置
type AppConfig struct {
	Name     string   `hcl:"name,label" json:"name"`
	LogLevel string   `hcl:"log_level,optional" json:"log_level,omitempty"`
	LogDir   string   `hcl:"log_dir,optional" json:"log_dir,omitempty"`
	Remain   hcl.Body `hcl:",remain"`
}

// ReadConfigFile 读取配置文件
func ReadConfigFile(options any) (*RootConfig, error) {
	var err error
	if settings == nil {
		settings = new(RootConfig)
		if verbose {
			fmt.Println("Config file is", cfgFile)
		}
		err = hclsimple.DecodeFile(cfgFile, nil, settings)
	}
	if err == nil && options != nil {
		gohcl.DecodeBody(settings.App.Remain, nil, options)
	}
	// 复制连接配置，用于同一个实例的不同数据库
	if len(settings.Repeats) > 0 {
		adds := dialect.RepeatConns(settings.Repeats, settings.Conns)
		if len(adds) > 0 {
			settings.Conns = append(settings.Conns, adds...)
		}
		settings.Repeats = []dialect.RepeatConfig{} //避免重复生成
	}
	return settings, err
}

// GetConfigLogger 获取配置中的Logger
func GetConfigLogger() (*utils.Logger, error) {
	if logger != nil {
		return logger, nil
	}
	var err error
	level, dir := "error", ""
	if settings, err = ReadConfigFile(nil); err == nil {
		app := settings.App
		level, dir = app.LogLevel, app.LogDir
	}
	return utils.NewLogger(level, dir), err
}
