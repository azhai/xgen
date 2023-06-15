package dialect

import (
	"fmt"
	"strconv"
	"strings"
)

const REDIS_DEFAULT_PORT uint16 = 6379

// Redis缓存
type Redis struct {
	Host     string `hcl:"host" json:"host"`
	Port     uint16 `hcl:"port,optional" json:"port,omitempty"`
	Database int    `hcl:"database,optional" json:"database,omitempty"`
}

// Name 驱动名
func (Redis) Name() string {
	return "redis"
}

// ImporterPath 驱动支持库
func (Redis) ImporterPath() string {
	return "github.com/gomodule/redigo/redis"
}

// IsXormDriver 是否Xorm支持的驱动
func (Redis) IsXormDriver() bool {
	return false
}

// QuoteIdent 字段或表名脱敏
func (Redis) QuoteIdent(ident string) string {
	return WrapWith(ident, "'", "'")
}

// ChangeDb 切换数据库
func (d *Redis) ChangeDb(database string) (bool, error) {
	db, err := strconv.Atoi(database)
	if err != nil {
		return false, err // 失败
	}
	d.Database = db
	return true, nil // 成功
}

// BuildDSN 生成DSN连接串
func (d Redis) BuildDSN() string {
	addr := DIALECT_DEFAULT_HOST
	if d.Host != "" {
		addr = GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("redis://%s/%d?", addr, d.Database)
	return dsn
}

// BuildFullDSN 生成带账号的完整DSN
func (d Redis) BuildFullDSN(username, password string) string {
	dsn, head := d.BuildDSN(), "redis://"
	if strings.HasPrefix(dsn, head) {
		account := GetAccount(username, password)
		dsn = head + account + "@" + dsn[len(head):]
	}
	return dsn
}
