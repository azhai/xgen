package main

import (
	"fmt"

	"github.com/azhai/xgen/models"
	// db "github.com/azhai/xgen/models/default"

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

	demo1()

	// var err error
	// fromid := 4705930
	// obj := new(db.CdrRecent)
	// opt1, opt2 := xq.WithTable(obj.TableName()), WithCdrUser(fromid)
	// bean := new(db.CdrRecentCore)
	// rit := xq.NewRowIterator(bean, "id", true, 15, 50)

	// var count int64
	// proc := func(bean any, col string) (int64, error) {
	// 	obj := bean.(*db.CdrRecentCore)
	// 	fmt.Println(">", obj.Id, obj.MsgTime)
	// 	return obj.Id, nil
	// }
	// count, err = rit.FindAll(db.Engine(), proc, opt1, opt2)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(count)

	// fmt.Println("--------------------------------------------------")

	// rch := xq.NewRowChannel(rit)
	// consume1 := func(val any) {
	// 	row := val.(map[string]any)
	// 	fmt.Println(row["id"], row["msg_time"])
	// }
	// rch.UpdateAll(db.Engine(), consume1, opt1, opt2)

	// fmt.Println("--------------------------------------------------")

	// consume2 := func(val any) {
	// 	ids := val.([]int64)
	// 	fmt.Println("<", ids)
	// }
	// err = rch.UpdateIndex(db.Engine(), consume2, "id", opt1, opt2)
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Println("执行完成。")
}

func demo1() {
	start, end, step := 1, 88, 10
	temp := start + step
	for temp < end {
		fmt.Println(start, "->", temp)
		start, temp = temp, temp+step
	}
	fmt.Println(start, "->", end)
}
