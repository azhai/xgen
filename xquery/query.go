package xquery

import (
	"fmt"
	"time"

	"xorm.io/xorm"
)

const (
	MaxReadSize  = 3000 // 一次读取最大行数
	MaxWriteSize = 200  // 一次读取最大行数
)

// BeanFunc 处理单行数据
type BeanFunc = func(bean any) (int64, error)

// FilterFunc 过滤查询
type FilterFunc = func(qr *xorm.Session) *xorm.Session

// ModifyFunc 修改操作，用于事务
type ModifyFunc = func(tx *xorm.Session) (int64, error)

// ScopeFunc 预置查询
type ScopeFunc = func(qr *xorm.Session, args ...any) *xorm.Session

// QueryOpts 查询附加条件
type QueryOpts struct {
	Bean   ITableName
	Filter FilterFunc
	Limit  int
	Order  string
	IsDesc bool
}

func (q QueryOpts) GetOrder() string {
	if q.Order == "" {
		return ""
	}
	if q.IsDesc {
		return q.Order + " DESC"
	} else {
		return q.Order + " ASC"
	}
}

func (q QueryOpts) Apply(query *xorm.Session) *xorm.Session {
	if q.Bean != nil {
		query = query.Table(q.Bean.TableName())
	}
	if q.Filter != nil {
		query = q.Filter(query)
	}
	if q.Limit > 0 {
		query = query.Limit(q.Limit)
	}
	if order := q.GetOrder(); order != "" {
		query = query.OrderBy(order)
	}
	return query
}

// Qprintf 对参数先进行转义Quote
func Qprintf(engine *xorm.Engine, format string, args ...any) string {
	if engine != nil {
		for i, arg := range args {
			args[i] = engine.Quote(arg.(string))
		}
	}
	return fmt.Sprintf(format, args...)
}

// ExecTx 执行事务
func ExecTx(engine *xorm.Engine, modify ModifyFunc) error {
	tx := engine.NewSession() // 必须是新的session
	defer tx.Close()
	_ = tx.Begin()
	if _, err := modify(tx); err != nil {
		_ = tx.Rollback() // 失败回滚
		return err
	}
	return tx.Commit()
}

// NegativeOffset 调整从后往前翻页
func NegativeOffset(offset, pagesize, total int) int {
	if remain := total % pagesize; remain > 0 {
		offset += pagesize - remain
	}
	return offset + total
}

// CalcPage 计算翻页
func CalcPage(pageno, pagesize, total int) (int, int) {
	if pagesize < 0 {
		return -1, 0
	} else if pagesize == 0 {
		return 0, 0
	}
	var offset int
	if pageno > 0 {
		offset = (pageno - 1) * pagesize
	} else if pageno < 0 && total > 0 {
		offset = NegativeOffset(pageno*pagesize, pagesize, total)
	}
	return pagesize, offset
}

// Paginate 分页查询
func Paginate(query *xorm.Session, pageno, pagesize int) *xorm.Session {
	var limit, offset int
	if pagesize > 0 && pageno < 0 {
		total, _ := query.Count()
		limit, offset = CalcPage(pageno, pagesize, int(total))
	} else {
		limit, offset = CalcPage(pageno, pagesize, 0)
	}
	if limit >= 0 {
		query = query.Limit(limit, offset)
	}
	return query
}

// Sequence 排序查询
func Sequence(query *xorm.Session, desc bool, args ...any) *xorm.Session {
	opts := QueryOpts{IsDesc: desc}
	if len(args) >= 1 && args[0] != nil {
		opts.Filter = func(qr *xorm.Session) *xorm.Session {
			return qr.Table(args[0])
		}
	}
	if len(args) >= 2 {
		if field, ok := args[1].(string); ok {
			opts.Order = field
		}
	}
	return opts.Apply(query)
}

// Recursive 递归查询
func Recursive(query *xorm.Session, opts QueryOpts, proc BeanFunc, msec int) (count int64, err error) {
	if count, err = opts.Apply(query).Count(); err != nil || count == 0 {
		return
	}
	if opts.Limit <= 0 && count > MaxReadSize {
		opts.Limit = MaxReadSize
	}
	// 递归查询
	id, rows := int64(0), new(xorm.Rows)
	query = opts.Apply(query)
	for err == nil {
		n := 0
		rows, err = query.Rows(opts.Bean)
		for rows.Next() {
			n++
			if err = rows.Scan(opts.Bean); err != nil {
				break
			}
			if id, err = proc(opts.Bean); err != nil {
				break
			}
		}
		err = rows.Close()
		if n == 0 { // 没有数据
			return count, err
		}

		// 下一次查询
		if msec > 0 {
			time.Sleep(time.Duration(msec) * time.Millisecond)
		}
		if opts.IsDesc {
			query = query.Where("id < ?", id)
		} else {
			query = query.Where("id > ?", id)
		}
	}
	return
}
