package dialect

import (
	"strings"
)

// SQLite3数据库
type Sqlite struct {
	Path string `hcl:"path,optional" json:"path"`
}

// Name 驱动名
func (Sqlite) Name() string {
	return "sqlite3"
}

// ImporterPath 驱动支持库
func (Sqlite) ImporterPath() string {
	return "github.com/mattn/go-sqlite3"
}

// IsXormDriver 是否Xorm支持的驱动
func (Sqlite) IsXormDriver() bool {
	return true
}

// QuoteIdent 字段或表名脱敏
func (Sqlite) QuoteIdent(ident string) string {
	return WrapWith(ident, "`", "`")
}

// ChangeDb 切换数据库
func (Sqlite) ChangeDb(database string) (bool, error) {
	return false, nil // 不支持
}

// BuildDSN 生成DSN连接串
func (d Sqlite) BuildDSN() string {
	if d.IsMemory() {
		d.Path = ":memory:"
	}
	return "file:" + d.Path + "?cache=shared&"
}

// BuildFullDSN 生成带账号的完整DSN
func (d Sqlite) BuildFullDSN(username, password string) string {
	dsn := d.BuildDSN()
	if !d.IsMemory() && username != "" {
		dsn += "_auth_user=" + username + "&"
		dsn += "_auth_pass=" + Escape(password) + "&"
	}
	return dsn
}

// IsMemory 是否内存数据库
func (d Sqlite) IsMemory() bool {
	return d.Path == "" || strings.ToLower(d.Path) == ":memory:"
}
