package xquery

import (
	"fmt"

	"xorm.io/xorm"
)

// FilterFunc 过滤查询
type FilterFunc = func(qr *xorm.Session) *xorm.Session

// ModifyFunc 修改操作，用于事务
type ModifyFunc = func(tx *xorm.Session) (int64, error)

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

// QueryAll 查询多行数据
func QueryAll(qr *xorm.Session, filter FilterFunc, pages ...int) *xorm.Session {
	if filter != nil {
		qr = filter(qr)
	}
	pageno, pagesize := 0, -1
	if len(pages) >= 1 {
		pageno = pages[0]
		if len(pages) >= 2 {
			pagesize = pages[1]
		}
	}
	return Paginate(qr, pageno, pagesize)
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

// Paginate 使用翻页
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
