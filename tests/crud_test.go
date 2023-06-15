package xgen_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	db "github.com/azhai/xgen/models/default"
	xq "github.com/azhai/xgen/xquery"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

// ------------------------------------------------------------
// MenuForTest 菜单
// ------------------------------------------------------------
type MenuForTest struct {
	Id              int `json:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	*xq.NestedMixin `json:",inline" xorm:"extends"`
	Path            string         `json:"path" xorm:"notnull default '' comment('路径') index VARCHAR(100)"`
	Title           string         `json:"title" xorm:"notnull default '' comment('名称') VARCHAR(50)"`
	Icon            string         `json:"icon" xorm:"comment('图标') VARCHAR(30)"`
	Remark          sql.NullString `json:"remark" xorm:"comment('说明备注') TEXT"`
	*xq.TimeMixin   `json:",inline" xorm:"extends"`
}

func (MenuForTest) TableName() string {
	return "x_menu_abcd"
}

// ------------------------------------------------------------
// the queries of MenuForTest
// ------------------------------------------------------------

// Load the queries of MenuForTest
func (m *MenuForTest) Load(opts ...xq.QueryOption) (bool, error) {
	opts = append(opts, xq.WithTable(m))
	return db.Query(opts...).Get(m)
}

// Save the queries of MenuForTest
func (m *MenuForTest) Save(changes map[string]any) error {
	return xq.ExecTx(db.Engine(), func(tx *xorm.Session) (int64, error) {
		if len(changes) == 0 {
			return tx.Table(m).Insert(m)
		} else if m.Id == 0 {
			changes["created_at"] = time.Now()
			return tx.Table(m).Insert(changes)
		} else {
			return tx.Table(m).ID(m.Id).Update(changes)
		}
	})
}

func Test01Create(t *testing.T) {
	err := db.Engine().Sync2(new(MenuForTest))
	assert.NoError(t, err)
}

func Test02Insert(t *testing.T) {
}

func Test03FindCount(t *testing.T) {
}

func Test04GetOne(t *testing.T) {
}

func Test05Delete(t *testing.T) {
}

func Test06Query(t *testing.T) {
	table := (MenuForTest{}).TableName()
	sql := fmt.Sprintf("DROP TABLE `%s`", table)
	_, err := db.Engine().Query(sql)
	assert.NoError(t, err)
}
