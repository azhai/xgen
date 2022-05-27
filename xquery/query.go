package xquery

import (
	"fmt"
	"time"

	"github.com/mitchellh/copystructure"
	"xorm.io/xorm"
)

const (
	MaxReadSize  = 3000 // 一次读取最大行数
	MaxWriteSize = 200  // 一次写入最大行数
)

// BeanFunc 处理单行数据
type BeanFunc func(bean any, col string) (int64, error)

// ModifyFunc 修改操作，用于事务
type ModifyFunc func(tx *xorm.Session) (int64, error)

// QueryOption 查询条件
type QueryOption func(qr *xorm.Session) *xorm.Session

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

// WithTable 限定数据表
func WithTable(tableOrBean any) QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		return qr.Table(tableOrBean)
	}
}

// ApplyOptions 使用查询条件
func ApplyOptions(qr *xorm.Session, opts []QueryOption) *xorm.Session {
	for _, opt := range opts {
		qr = opt(qr)
	}
	return qr
}

// WithWhere where查询
func WithWhere(cond string, args ...any) QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		return qr.Where(cond, args...)
	}
}

// WithRange in查询
func WithRange(col string, args ...any) QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		return qr.In(col, args...)
	}
}

// WithOrderBy 限定排序
func WithOrderBy(column string, desc bool) QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		orient := " ASC"
		if desc {
			orient = " DESC"
		}
		return qr.OrderBy(column + orient)
	}
}

// WithLimit 限定最大行数
func WithLimit(limit int, offset ...int) QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		return qr.Limit(limit, offset...)
	}
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

// WithPage 分页查询
func WithPage(pageno, pagesize int) QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		var limit, offset int
		if pagesize > 0 && pageno < 0 {
			total, _ := qr.Count()
			limit, offset = CalcPage(pageno, pagesize, int(total))
		} else {
			limit, offset = CalcPage(pageno, pagesize, 0)
		}
		if limit >= 0 {
			qr = qr.Limit(limit, offset)
		}
		return qr
	}
}

// Recursion 递归查询
type Recursion struct {
	orderCol string
	isDesc   bool
	pageSize int           // 单次查询最大行数
	sleepGap time.Duration // 两次查询间的休眠时长
	Bean     any
}

// NewRecursion 创建递归查询
func NewRecursion(bean any, order string, desc bool, size, msec int) *Recursion {
	gap := time.Duration(msec) * time.Millisecond
	return &Recursion{
		Bean:     bean,
		orderCol: order, isDesc: desc,
		pageSize: size, sleepGap: gap,
	}
}

// IsEnough 没有更多行需要查询了
func (r Recursion) IsEnough(n int) bool {
	if n <= 0 {
		return true
	}
	if r.pageSize > 0 && n < r.pageSize {
		return true
	}
	return false
}

// All 递归查询
func (r Recursion) All(eng *xorm.Engine, proc BeanFunc,
	opts ...QueryOption) (count int64, err error) {
	if model, ok := r.Bean.(ITableName); ok {
		table := model.TableName()
		opts = append(opts, WithTable(table))
		if r.orderCol == "" {
			r.orderCol = GetPrimarykey(eng, model).Name
		}
	}
	// 查询符合条件的总行数
	query := eng.NewSession()
	count, err = ApplyOptions(query, opts).Count()
	if err != nil || count == 0 {
		return
	}
	if r.pageSize <= 0 && count > MaxReadSize {
		r.pageSize = MaxReadSize
	}
	opts = append(opts, WithLimit(r.pageSize))
	opts = append(opts, WithOrderBy(r.orderCol, r.isDesc))

	// 递归查询
	id, rows := int64(0), new(xorm.Rows)
	bean, _ := copystructure.Copy(r.Bean)
	for err == nil {
		n := 0
		rows, err = ApplyOptions(query, opts).Rows(r.Bean)
		for rows.Next() {
			n++
			if err = rows.Scan(bean); err != nil {
				break
			}
			if id, err = proc(bean, r.orderCol); err != nil {
				break
			}
		}
		err = rows.Close()
		if r.IsEnough(n) { // 没有更多数据
			return count, err
		}

		// 下一次查询
		if r.sleepGap > 0 {
			time.Sleep(r.sleepGap)
		}
		if r.isDesc {
			query = query.Where(r.orderCol+" < ?", id)
		} else {
			query = query.Where(r.orderCol+" > ?", id)
		}
	}
	return
}
