package dialect

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/azhai/gozzo/logging/adapters/xormlog"
	"github.com/azhai/xgen/utils"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"xorm.io/xorm"
)

const DIALECT_DEFAULT_HOST = "127.0.0.1"

var (
	ConcatWith = utils.ConcatWith
	WrapWith   = utils.WrapWith
	Escape     = url.QueryEscape
)

// Dialect 不同数据库的驱动配置
type Dialect interface {
	Name() string                                  // 驱动名
	ImporterPath() string                          // 驱动支持库
	IsXormDriver() bool                            // 是否Xorm支持的驱动
	QuoteIdent(ident string) string                // 字段或表名脱敏
	ChangeDb(database string) (bool, error)        // 切换数据库
	BuildDSN() string                              // 生成DSN连接串
	BuildFullDSN(username, password string) string // 生成带账号的完整DSN
}

// CreateDialectByName 根据名称创建驱动配置
func CreateDialectByName(name string) Dialect {
	name = strings.ToLower(name)
	switch name {
	default:
		return nil
	case "flashdb":
		return &FlashDB{}
	case "mariadb", "mysql":
		return &Mysql{}
	case "pgsql", "postgres":
		return &Postgres{}
	case "redis":
		return &Redis{}
	case "sqlite", "sqlite3":
		return &Sqlite{}
	}
}

// RepeatConfig 复制连接参数，只有数据库不同，目前只支持Mysql/Postgres/Redis
type RepeatConfig struct {
	Type     string   `hcl:"type,label" json:"type"`
	Key      string   `hcl:"key" json:"key"`
	DbPrefix string   `hcl:"db_prefix,optional" json:"db_prefix,omitempty"`
	DbNames  []string `hcl:"db_names,optional" json:"db_names,omitempty"`
}

// RepeatConns 复制配置
func RepeatConns(reps []RepeatConfig, conns []ConnConfig) (adds []ConnConfig) {
	var rep RepeatConfig
	repeatDict := make(map[string]RepeatConfig)
	for _, rep = range reps {
		repeatDict[rep.Type] = rep
	}
	for _, cfg := range conns {
		ok := false
		if rep, ok = repeatDict[cfg.Type]; !ok || cfg.Key != rep.Key {
			continue
		}
		for _, name := range rep.DbNames {
			newbie := cfg.CopyIt(name)
			dia := newbie.LoadDialect()
			if ok, _ = dia.ChangeDb(rep.DbPrefix + name); !ok {
				break
			}
			// fmt.Printf("addrs: %p %p %p\n", dia, cfg.LoadDialect(), newbie.LoadDialect())
			// fmt.Printf("addrs: %s %s\n", cfg.Key, newbie.Key)
			adds = append(adds, newbie)
		}
	}
	return
}

// ConnConfig 连接配置
type ConnConfig struct {
	Type     string     `hcl:"type,label" json:"type"` // 数据库类型
	Key      string     `hcl:"key,label" json:"key"`   // 数据库连接名
	LogFile  string     `hcl:"log_file,optional" json:"log_file,omitempty"`
	Username string     `hcl:"username,optional" json:"username,omitempty"`
	Password string     `hcl:"password,optional" json:"password,omitempty"`
	DSN      string     `hcl:"dsn,optional" json:"dsn,omitempty"`
	Options  url.Values `hcl:"options,optional" json:"options,omitempty"`
	Remain   hcl.Body   `hcl:",remain"`
	Dialect  Dialect
}

// LoadDialect 加载数据库驱动配置
func (c *ConnConfig) LoadDialect() Dialect {
	if c.Type == "" || c.Dialect != nil {
		return c.Dialect
	}
	c.Dialect = CreateDialectByName(c.Type)
	if c.Dialect != nil && c.Remain != nil {
		gohcl.DecodeBody(c.Remain, nil, c.Dialect)
	}
	return c.Dialect
}

// CopyIt 复制一份完整配置，但修改它的连接名
func (c ConnConfig) CopyIt(key string) ConnConfig {
	c.LoadDialect()
	c.Key = key
	return c
}

// Name 数据库驱动名
func (c ConnConfig) Name() string {
	if d := c.LoadDialect(); d != nil {
		return d.Name()
	}
	return c.Type
}

// GetDSN 获取DSN连接串，可选是否带账号密码
func (c ConnConfig) GetDSN(full bool) string {
	var dsn string
	if d := c.LoadDialect(); d != nil {
		if c.DSN == "" {
			c.DSN = d.BuildDSN()
		}
		if full {
			dsn = d.BuildFullDSN(c.Username, c.Password)
		}
	}
	if dsn == "" {
		dsn = c.DSN
	}
	if args := c.Options.Encode(); args != "" {
		dsn += args
	}
	return strings.TrimRight(dsn, " ?&")
}

// QuickConnect 连接数据库，需要先导入对应驱动
func (c ConnConfig) QuickConnect(logsql, verbose bool) *xorm.Engine {
	engine, err := xorm.NewEngine(c.Name(), c.GetDSN(true))
	if verbose && err != nil {
		panic(err)
	}
	if logfile := c.LogFile; logfile != "" && logsql {
		engine.SetLogger(xormlog.NewLogger(logfile))
	}
	return engine
}

// GetAddr 获得数据库的连接地址和端口，用于TCP协议的数据库连接
func GetAddr(host string, port uint16) string {
	if port == 0 {
		return host
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// GetAccount 获得账号密码，用于DSN连接串
func GetAccount(username, password string) string {
	if username == "" {
		return password
	}
	if password == "" {
		return username
	}
	return fmt.Sprintf("%s:%s", username, password)
}

// // ConnectXorm 连接数据库
// func ConnectXorm(cfg ConnConfig) *xorm.Engine {
// 	if d := cfg.LoadDialect(); d == nil || !d.IsXormDriver() {
// 		return nil
// 	}
// 	return cfg.QuickConnect(true, true)
// }

// // ConnectRedis 连接数据库
// func ConnectRedis(cfg ConnConfig, db int) *redisw.RedisWrapper {
// 	if cfg.Type != "redis" {
// 		return nil
// 	}
// 	conn, err := redisw.NewRedisConnDb(cfg, db)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return redisw.NewRedisConnMux(conn, nil)
// }

// // ConnectFlashDB 连接数据库
// func ConnectFlashDB(cfg ConnConfig) *flashdb.FlashDB {
// 	if cfg.Type != "flashdb" {
// 		return nil
// 	}
// 	config := &flashdb.Config{
// 		Path: cfg.Path, EvictionInterval: cfg.EvictionInterval,
// 	}
// 	conn, err := flashdb.New(config)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return conn
// }
