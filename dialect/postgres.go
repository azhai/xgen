package dialect

import (
	"fmt"
	"strings"
)

const PGSQL_DEFAULT_PORT uint16 = 5432

// PostgreSQL数据库
type Postgres struct {
	Host     string `hcl:"host" json:"host"`
	Port     uint16 `hcl:"port,optional" json:"port,omitempty"`
	Database string `hcl:"database,optional" json:"database,omitempty"`
}

// Name 驱动名
func (Postgres) Name() string {
	return "postgres"
}

// ImporterPath 驱动支持库
func (Postgres) ImporterPath() string {
	return "github.com/lib/pq"
}

// IsXormDriver 是否Xorm支持的驱动
func (Postgres) IsXormDriver() bool {
	return true
}

// QuoteIdent 字段或表名脱敏
func (Postgres) QuoteIdent(ident string) string {
	return WrapWith(ident, `"`, `"`)
}

// ChangeDb 切换数据库
func (d *Postgres) ChangeDb(database string) (bool, error) {
	d.Database = database
	return true, nil // 成功
}

// BuildDSN 生成DSN连接串
func (d Postgres) BuildDSN() string {
	addr := DIALECT_DEFAULT_HOST
	if d.Host != "" {
		addr = GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("postgres://%s/%s?", addr, d.Database)
	return dsn
}

// BuildFullDSN 生成带账号的完整DSN
func (d Postgres) BuildFullDSN(username, password string) string {
	dsn, head := d.BuildDSN(), "postgres://"
	if strings.HasPrefix(dsn, head) {
		account := GetAccount(username, password)
		dsn = head + account + "@" + dsn[len(head):]
	}
	return dsn
}
