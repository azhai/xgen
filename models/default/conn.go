package db

import (
	"fmt"

	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/models"
	xq "github.com/azhai/xgen/xquery"
	_ "github.com/lib/pq"
	"xorm.io/xorm"
)

var engine *xorm.Engine

// ConnectXorm 连接数据库
func ConnectXorm(cfg dialect.ConnConfig) *xorm.Engine {
	if d := cfg.LoadDialect(); d == nil || !d.IsXormDriver() {
		return nil
	}
	return cfg.QuickConnect(true, true)
}

// Engine 获取当前数据库连接
func Engine() *xorm.Engine {
	if engine == nil {
		cfg := models.GetConnConfig("default")
		engine = ConnectXorm(cfg)
		_ = SyncModels(engine)
	}
	return engine
}

// Query 生成查询
func Query(opts ...xq.QueryOption) *xorm.Session {
	qr := Engine().NewSession()
	if len(opts) > 0 {
		return xq.ApplyOptions(qr, opts)
	}
	return qr
}

// Quote 转义表名或字段名
func Quote(value string) string {
	return Engine().Quote(value)
}

// InsertBatch 写入多行数据
func InsertBatch(tableName string, rows ...any) error {
	if len(rows) == 0 {
		return nil
	}
	modify := func(tx *xorm.Session) (int64, error) {
		return tx.Table(tableName).Insert(rows)
	}
	return xq.ExecTx(Engine(), modify)
}

// UpdateBatch 更新多行数据
func UpdateBatch(tableName, pkey string, ids any, changes map[string]any) error {
	if len(changes) == 0 || ids == nil {
		return nil
	}
	modify := func(tx *xorm.Session) (int64, error) {
		return tx.Table(tableName).In(pkey, ids).Update(changes)
	}
	return xq.ExecTx(Engine(), modify)
}

// SyncModels 同步数据库表结构
func SyncModels(eng *xorm.Engine) error {
	if eng == nil {
		return fmt.Errorf("the connection is lost")
	}
	return eng.Sync()
}
