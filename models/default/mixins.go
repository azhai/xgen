package db

// ------------------------------------------------------------
// UserProfileCore 用户资料表
// ------------------------------------------------------------
type UserProfileCore struct {
	Id       int    `json:"id" xorm:"notnull pk autoincr comment('自增ID') UNSIGNED INT(10)"`
	Username string `json:"username" xorm:"notnull default '' comment('用户名') index(username) VARCHAR(20)"`
	Gender   int    `json:"gender" xorm:"notnull default 0 comment('性别 0未知 1男 2女') UNSIGNED TINYINT(1)"`
	IsDel    bool   `json:"is_del" xorm:"notnull default 0 comment('是否删除 0正常 1已删除') UNSIGNED TINYINT(1)"`
}
