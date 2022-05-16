package db

import (
	"github.com/azhai/xgen/models"

	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/xquery"
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

var (
	engine *xorm.Engine
)

// ConnectXorm 连接数据库
func ConnectXorm(cfg dialect.ConnConfig) *xorm.Engine {
	if d := cfg.LoadDialect(); d == nil || !d.IsXormDriver() {
		return nil
	}
	engine := cfg.QuickConnect(true, true)
	if cfg.LogFile != "" {
		logger := xquery.NewSqlLogger(cfg.LogFile)
		engine.SetLogger(logger)
	}
	return engine
}

// Engine 获取当前数据库连接
func Engine() *xorm.Engine {
	if engine == nil {
		cfg := models.GetConnConfig("default")
		engine = ConnectXorm(cfg)
	}
	return engine
}

// Quote 转义表名或字段名
func Quote(value string) string {
	return Engine().Quote(value)
}

// Table 查询某张数据表
func Table(args ...interface{}) *xorm.Session {
	qr := Engine().NewSession()
	if args == nil {
		return qr
	}
	return qr.Table(args[0])
}

// InsertBatch 写入多行数据
func InsertBatch(tableName string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	return xquery.ExecTx(Engine(), func(tx *xorm.Session) (int64, error) {
		return tx.Table(tableName).Insert(rows)
	})
}
