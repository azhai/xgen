package xquery

import (
	"fmt"
	"time"

	xutils "github.com/azhai/xgen/utils"
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

// ApplyOptions 使用查询条件
func ApplyOptions(qr *xorm.Session, opts []QueryOption) *xorm.Session {
	for _, opt := range opts {
		qr = opt(qr)
	}
	return qr
}

// WithTable 限定数据表和字段
func WithTable(tableOrBean any, cols ...string) QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		qr = qr.Table(tableOrBean)
		if len(cols) > 0 {
			qr = qr.Cols(cols...)
		}
		return qr
	}
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
		if column == "" {
			return qr
		}
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
		if limit <= 0 {
			return qr
		}
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

// RowIterator 迭代查询
type RowIterator struct {
	orderCol string
	isDesc   bool
	pageSize int           // 单次查询最大行数
	sleepGap time.Duration // 两次查询间的休眠时长
	Bean     any
}

// NewRowIterator 创建迭代查询
func NewRowIterator(bean any, order string, desc bool, size, msec int) *RowIterator {
	if size <= 0 {
		size = MaxReadSize
	}
	gap := time.Duration(msec) * time.Millisecond
	return &RowIterator{
		Bean:     bean,
		orderCol: order, isDesc: desc,
		pageSize: size, sleepGap: gap,
	}
}

// IsEnough 没有更多行需要查询了
func (r *RowIterator) IsEnough(n int) bool {
	if n <= 0 {
		return true
	}
	if r.pageSize > 0 && n < r.pageSize {
		return true
	}
	return false
}

// prepare 准备查询范围条件
func (r *RowIterator) prepare(eng *xorm.Engine, opts []QueryOption) []QueryOption {
	if model, ok := r.Bean.(ITableName); ok {
		table := model.TableName()
		opts = append(opts, WithTable(table))
		if r.orderCol == "" {
			r.orderCol = GetPrimaryKey(eng, model).Name
		}
	}
	// 查询符合条件的总行数
	opts = append(opts, WithLimit(r.pageSize), WithOrderBy(r.orderCol, r.isDesc))
	return opts
}

// IterBean 迭代查询对象
func (r *RowIterator) IterBean(qr *xorm.Session, proc BeanFunc,
	opts []QueryOption,
) (count int64, err error) {
	rows := new(xorm.Rows)
	for {
		id, n := int64(0), 0
		rows, err = ApplyOptions(qr, opts).Rows(r.Bean)
		for rows.Next() {
			n++
			bean, _ := copystructure.Copy(r.Bean)
			if err = rows.Scan(bean); err != nil || bean == nil {
				break
			}
			if id, err = proc(bean, r.orderCol); err != nil {
				break
			}
		}
		// 计数，关闭本次结果集
		count += int64(n)
		if errClose := rows.Close(); errClose != nil {
			err = errClose
		}
		if err != nil || r.IsEnough(n) { // 没有更多数据
			return
		}
		// 下一次查询
		if r.sleepGap > 0 {
			time.Sleep(r.sleepGap)
		}
		if r.isDesc {
			qr = qr.Where(r.orderCol+" < ?", id)
		} else {
			qr = qr.Where(r.orderCol+" > ?", id)
		}
	}
	return
}

// IterCol 迭代查询主键
func (r *RowIterator) IterCol(qr *xorm.Session, proc BeanFunc,
	col string, opts []QueryOption,
) (count int64, err error) {
	if col == "" {
		col = r.orderCol
	}
	var lst []int64
	id, n := int64(0), 0
	for err == nil {
		err = ApplyOptions(qr, opts).Cols(col).Find(&lst)
		if n = len(lst); n > 0 {
			_, err = proc(lst, col)
			count += int64(n)
		}
		if n == 0 || r.IsEnough(n) { // 没有更多数据
			return
		}
		id = lst[n-1]
		lst = lst[:0] // 清空但不回收容量
		// 下一次查询
		if r.sleepGap > 0 {
			time.Sleep(r.sleepGap)
		}
		if r.isDesc {
			qr = qr.Where(r.orderCol+" < ?", id)
		} else {
			qr = qr.Where(r.orderCol+" > ?", id)
		}
	}
	return
}

// FindIndex 迭代查询主键列表，后累计总行数
func (r *RowIterator) FindIndex(eng *xorm.Engine, proc BeanFunc,
	col string, opts ...QueryOption,
) (count int64, err error) {
	opts = r.prepare(eng, opts)
	qr := eng.NewSession()
	count, err = r.IterCol(qr, proc, col, opts)
	return
}

// FindAll 迭代查询每行，后累计总行数
func (r *RowIterator) FindAll(eng *xorm.Engine, proc BeanFunc,
	opts ...QueryOption,
) (count int64, err error) {
	opts = r.prepare(eng, opts)
	qr := eng.NewSession()
	count, err = r.IterBean(qr, proc, opts)
	return
}

// FindCount 迭代查询，先查询总行数
func (r *RowIterator) FindCount(eng *xorm.Engine, proc BeanFunc,
	opts ...QueryOption,
) (count int64, err error) {
	opts = r.prepare(eng, opts)
	qr := eng.NewSession()
	count, err = ApplyOptions(qr, opts).Count()
	if err != nil || count == 0 { // 没有符合条件的数据行
		return
	}
	_, err = r.IterBean(qr, proc, opts)
	return
}

// RowChannel 队列异步操作
type RowChannel struct {
	dataCh chan any
	*RowIterator
}

// NewRowChannel 创建迭代查询
func NewRowChannel(iter *RowIterator) *RowChannel {
	return &RowChannel{RowIterator: iter}
}

// Update 通用队列生产和消费
func (r *RowChannel) Update(eng *xorm.Engine, proc BeanFunc,
	consume func(val any), col string, opts []QueryOption,
) error {
	r.dataCh = make(chan any)
	errCh := make(chan error, 1)  // 遇到一个错误就返回
	go func(errCh chan<- error) { // 生产者放入协程，消费者才不会漏掉最后一个元素
		var err error
		if col == "" {
			_, err = r.FindAll(eng, proc, opts...)
		} else {
			_, err = r.FindIndex(eng, proc, col, opts...)
		}
		errCh <- err // 不用判断nil
		close(r.dataCh)
	}(errCh)
	for val := range r.dataCh {
		consume(val)
	}
	return <-errCh
}

// UpdateIndex 消费主键列表
func (r *RowChannel) UpdateIndex(eng *xorm.Engine, consume func(val any),
	col string, opts ...QueryOption,
) error {
	if col == "" {
		col = r.RowIterator.orderCol
	}
	proc := func(bean any, col string) (int64, error) {
		r.dataCh <- bean
		return 0, nil
	}
	return r.Update(eng, proc, consume, col, opts)
}

// UpdateAll 消费每行字典
func (r *RowChannel) UpdateAll(eng *xorm.Engine, consume func(val any),
	opts ...QueryOption,
) error {
	proc := func(bean any, col string) (int64, error) {
		row, err := xutils.Obj2Dict(bean)
		if err != nil {
			return 0, err
		}
		r.dataCh <- row
		if id, ok := row[col]; ok {
			return xutils.JsonInt64(id), nil
		}
		return 0, err
	}
	return r.Update(eng, proc, consume, "", opts)
}

// UpdateSlice 批量更新
func UpdateSlice[T any](dataCh <-chan T, size int, batch func(lst []T) error) (err error) {
	if size <= 0 {
		size = MaxWriteSize
	}
	var lst []T
	for val := range dataCh {
		lst = append(lst, val)
		if len(lst) >= size { // 凑足数量就处理一批
			err = batch(lst)
			lst = lst[:0]
		}
	}
	if len(lst) > 0 { // 最后一批
		err = batch(lst)
	}
	return
}
