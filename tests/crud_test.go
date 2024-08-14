package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	db "github.com/azhai/xgen/models/default"
	"github.com/azhai/xgen/xquery"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

// ------------------------------------------------------------
// MenuForTest 菜单
// ------------------------------------------------------------
type MenuForTest struct {
	Id                  int `json:"id" xorm:"notnull pk autoincr UNSIGNED INT(10)"`
	*xquery.NestedMixin `json:",inline" xorm:"extends"`
	Path                string         `json:"path" xorm:"notnull default '' comment('路径') index VARCHAR(100)"`
	Title               string         `json:"title" xorm:"notnull default '' comment('名称') VARCHAR(50)"`
	Icon                string         `json:"icon" xorm:"comment('图标') VARCHAR(30)"`
	Remark              sql.NullString `json:"remark" xorm:"comment('说明备注') TEXT"`
	*xquery.TimeMixin   `json:",inline" xorm:"extends"`
}

func (*MenuForTest) TableName() string {
	return "x_menu_abcd"
}

// ------------------------------------------------------------
// the queries of MenuForTest
// ------------------------------------------------------------

func (m *MenuForTest) Load(where any, args ...any) (bool, error) {
	return db.Engine().Table(m).Where(where, args...).Get(m)
}

func (m *MenuForTest) Save(changes map[string]any) error {
	return xquery.ExecTx(db.Engine(), func(tx *xorm.Session) (int64, error) {
		if changes == nil || m.Id == 0 {
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
	table := new(MenuForTest).TableName()
	sql := fmt.Sprintf("DROP TABLE `%s`", table)
	_, err := db.Engine().Query(sql)
	assert.NoError(t, err)
}
