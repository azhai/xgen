package dialect

import (
	"fmt"
)

const MYSQL_DEFAULT_PORT uint16 = 3306

// MySQL或MariaDB数据库
type Mysql struct {
	Host     string `hcl:"host" json:"host"`
	Port     uint16 `hcl:"port,optional" json:"port,omitempty"`
	Database string `hcl:"database,optional" json:"database,omitempty"`
}

// Name 驱动名
func (Mysql) Name() string {
	return "mysql"
}

// ImporterPath 驱动支持库
func (Mysql) ImporterPath() string {
	return "github.com/go-sql-driver/mysql"
}

// IsXormDriver 是否Xorm支持的驱动
func (Mysql) IsXormDriver() bool {
	return true
}

// QuoteIdent 字段或表名脱敏
func (Mysql) QuoteIdent(ident string) string {
	return WrapWith(ident, "`", "`")
}

// ChangeDb 切换数据库
func (d *Mysql) ChangeDb(database string) (bool, error) {
	d.Database = database
	return true, nil // 成功
}

// BuildDSN 生成DSN连接串
func (d Mysql) BuildDSN() string {
	addr := DIALECT_DEFAULT_HOST
	if d.Host != "" {
		addr = GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("tcp(%s)/%s", addr, d.Database)
	dsn += "?parseTime=true&loc=Local&"
	return dsn
}

// BuildFullDSN 生成带账号的完整DSN
func (d Mysql) BuildFullDSN(username, password string) string {
	dsn := d.BuildDSN()
	if dsn != "" {
		account := GetAccount(username, password)
		dsn = account + "@" + dsn
	}
	return dsn
}
