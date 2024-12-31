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

// GetDbSettings 读取默认配置文件
func GetDbSettings(root *config.RootConfig) (*DbSettings, error) {
	settings := new(DbSettings)
	err := root.ParseRemain(settings)
	return settings, err
}

// GetConns 读取默认配置文件
func (s *DbSettings) GetConns() []dialect.ConnConfig {
	// 复制连接配置，用于同一个实例的不同数据库
	if len(s.Repeats) > 0 {
		adds := dialect.RepeatConns(s.Repeats, s.Conns)
		if len(adds) > 0 {
			s.Conns = append(s.Conns, adds...)
		}
		s.Repeats = []dialect.RepeatConfig{} // 避免重复生成
	}
	return s.Conns
}
