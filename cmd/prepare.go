package cmd

import (
	"github.com/azhai/gozzo/config"
	reverse "github.com/azhai/xgen"
	"github.com/azhai/xgen/dialect"
)

var dbSettings = new(DbSettings)

// DbSettings 数据库、缓存相关配置
type DbSettings struct {
	Reverse *reverse.ReverseConfig `hcl:"reverse,block" json:"reverse,omitempty"`
	Repeats []dialect.RepeatConfig `hcl:"repeat,block" json:"repeat"`
	Conns   []dialect.ConnConfig   `hcl:"conn,block" json:"conn"`
}

// LoadConfigFile 读取默认配置文件
func LoadConfigFile(reload bool) (*config.RootConfig, error) {
	if reload == false { //  不重复解析配置文件
		theSettings := config.GetTheSettings()
		if theSettings != nil {
			return theSettings, nil
		}
	}
	theSettings, err := config.ParseConfigFile(dbSettings)
	// 复制连接配置，用于同一个实例的不同数据库
	if len(dbSettings.Repeats) > 0 {
		adds := dialect.RepeatConns(dbSettings.Repeats, dbSettings.Conns)
		if len(adds) > 0 {
			dbSettings.Conns = append(dbSettings.Conns, adds...)
		}
		dbSettings.Repeats = []dialect.RepeatConfig{} // 避免重复生成
	}
	return theSettings, err
}

// GetDbSettings 返回数据库配置单例
func GetDbSettings() *DbSettings {
	return dbSettings
}

// GetConnConfigs 返回数据库连接
func GetConnConfigs() (conns []dialect.ConnConfig) {
	if dbSettings != nil {
		conns = dbSettings.Conns
	}
	return
}
