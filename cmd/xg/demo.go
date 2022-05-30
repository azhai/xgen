package main

import (
	"fmt"

	"github.com/azhai/xgen/models"

	// db "github.com/azhai/xgen/models/default"
	// "github.com/azhai/xgen/models/repo"
	// "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
	"xorm.io/xorm"
)

// WithCdrUser 有此用户参与的通话
func WithCdrUser(userid int) xq.QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		cond := "fromid = ? OR toid = ?"
		return qr.Where(cond, userid, userid)
	}
}

// WithCdrPair 两个用户之间的通话
func WithCdrPair(fromid, toid int) xq.QueryOption {
	return func(qr *xorm.Session) *xorm.Session {
		cond := "fromid = ? AND toid = ? OR toid = ? AND fromid = ?"
		return qr.Where(cond, fromid, toid, fromid, toid)
	}
}

func runTheDemo() {
	models.Setup()

	// var row = map[string]any{}
	// qr := repo.Query().Table("repository").Asc("id")
	// if _, err := qr.Get(&row); err != nil {
	// 	fmt.Println(err)
	// }
	// utils.PrintJson(row)

	// var err error
	// obj := new(db.CdrRecent)
	// opt := xq.WithOrderBy("id", true)
	// if _, err = obj.Load(opt); err != nil {
	// 	panic(err)
	// }
	// utils.PrintJson(obj)

	// opt1 := xq.WithTable(obj.TableName())
	// opt2 := WithCdrUser(obj.Fromid)
	// proc := func(bean any, col string) (int64, error) {
	// 	obj := bean.(*db.CdrRecentCore)
	// 	fmt.Print(obj.Id, " >> ")
	// 	utils.PrintJson(obj)
	// 	return int64(obj.Id), nil
	// }
	// bean, count := new(db.CdrRecentCore), int64(0)
	// rec := xq.NewRecursion(bean, "id", true, 100, 100)
	// count, err = rec.All(db.Engine(), proc, opt1, opt2)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(count)

	fmt.Println("执行完成。")
}
