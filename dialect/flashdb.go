package dialect

import (
	"net/url"
	"strconv"
)

const FLASHDB_DEFAULT_PORT uint16 = 8000

// FlashDB 一个golang写的类似redis的缓存
type FlashDB struct {
	Path             string `hcl:"path,optional" json:"path,omitempty"`
	EvictionInterval int    `hcl:"eviction_interval,optional" json:"eviction_interval,omitempty"`
}

// Name 驱动名
func (FlashDB) Name() string {
	return "flashdb"
}

// ImporterPath 驱动支持库
func (FlashDB) ImporterPath() string {
	return "github.com/arriqaaq/flashdb"
}

// IsXormDriver 是否Xorm支持的驱动
func (FlashDB) IsXormDriver() bool {
	return false
}

// QuoteIdent 字段或表名脱敏
func (FlashDB) QuoteIdent(ident string) string {
	return WrapWith(ident, "'", "'")
}

// ChangeDb 切换数据库
func (FlashDB) ChangeDb(database string) (bool, error) {
	return false, nil // 不支持
}

// BuildDSN 生成DSN连接串
func (d FlashDB) BuildDSN() string {
	data := url.Values{}
	data.Set("path", d.Path)
	if d.EvictionInterval > 0 {
		data.Set("eviction_interval", strconv.Itoa(d.EvictionInterval))
	}
	return data.Encode()
}

// BuildFullDSN 生成带账号的完整DSN
func (d FlashDB) BuildFullDSN(username, password string) string {
	return d.BuildDSN()
}
