package xquery

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// TimeMixin 时间相关的三个典型字段
type TimeMixin struct {
	CreatedAt time.Time `json:"created_at" xorm:"created comment('创建时间') TIMESTAMP"`       // 创建时间
	UpdatedAt time.Time `json:"updated_at" xorm:"updated comment('更新时间') TIMESTAMP"`       // 更新时间
	DeletedAt time.Time `json:"deleted_at" xorm:"deleted comment('删除时间') index TIMESTAMP"` // 删除时间
}

// ITableName 数据表名
type ITableName interface {
	TableName() string
}

// ITableComment 数据表注释
type ITableComment interface {
	TableComment() string
}

// NewNullString string 与 NullString 相互转换
func NewNullString(word string) sql.NullString {
	return sql.NullString{String: word, Valid: word != ""}
}

func GetNullString(data sql.NullString) (word string) {
	if data.Valid {
		word = data.String
	}
	return
}

// FindTables 找出符合前缀的表名
func FindTables(engine *xorm.Engine, prefix string, fullName bool) []string {
	var result []string
	db, ctx := engine.DB(), context.Background()
	tables, err := engine.Dialect().GetTables(db, ctx)
	if err != nil {
		return result
	}
	prelen := len(prefix)
	for _, t := range tables {
		if prelen > 0 && !strings.HasPrefix(t.Name, prefix) {
			continue
		}
		if fullName {
			result = append(result, t.Name)
		} else {
			result = append(result, t.Name[prelen:])
		}
	}
	return result
}

// CreateTableLike 复制表结构，只用于MySQL
func CreateTableLike(engine *xorm.Engine, curr, orig string) (bool, error) {
	if engine.DriverName() != "mysql" {
		err := fmt.Errorf("only support mysql/mariadb database !")
		return false, err
	}
	exists, err := engine.IsTableExist(curr)
	if err != nil || exists {
		return false, err
	}
	sql := "CREATE TABLE IF NOT EXISTS %s LIKE %s"
	_, err = engine.Exec(Qprintf(engine, sql, curr, orig))
	return err == nil, err
}

// GetPrimarykey 获取Model的主键
func GetPrimarykey(engine *xorm.Engine, m any) *schemas.Column {
	table, err := engine.TableInfo(m)
	if err != nil {
		return nil
	}
	if cols := table.PKColumns(); len(cols) > 0 {
		return cols[0]
	}
	return nil
}
