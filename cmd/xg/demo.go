package main

import (
	"fmt"

	"github.com/azhai/xgen/models"
	db "github.com/azhai/xgen/models/default"
	"github.com/azhai/xgen/xquery"
	jsoniter "github.com/json-iterator/go"
	"xorm.io/xorm"
)

var (
	json  = jsoniter.ConfigCompatibleWithStandardLibrary
	print = func(data any) (err error) {
		var body []byte
		if body, err = json.Marshal(data); err == nil {
			fmt.Println(string(body))
		}
		return
	}
)

func runTheDemo() {
	models.Setup()
	db.AddScope("@cdr-user", func(qr *xorm.Session, args ...any) *xorm.Session {
		if len(args) >= 2 {
			cond := "fromid = ? AND toid = ? OR toid = ? AND fromid = ?"
			fromid, toid := args[0], args[1]
			return qr.Where(cond, fromid, toid, fromid, toid)
		} else if len(args) == 1 {
			cond, userid := "fromid = ? OR toid = ?", args[0]
			return qr.Where(cond, userid, userid)
		}
		return qr.Limit(1000)
	})

	var err error
	obj := new(db.CdrRecent)
	query := db.Scope(db.Table(obj), "@last", obj, "id")
	if _, err = query.Get(obj); err != nil {
		panic(err)
	}
	fmt.Println(query.LastSQL())
	print(obj)

	// var rows []*db.CdrRecent
	// count := int64(0)
	// query = db.Scope(db.Table(obj), "@cdr-user", obj.Fromid, obj.Toid)
	// if count, err = query.FindAndCount(&rows); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("there are N rows, N =", count)
	// for i, row := range rows {
	// 	fmt.Print(i+1, " > ")
	// 	print(row)
	// }

	count := int64(0)
	opts := xquery.QueryOpts{
		Bean: new(db.CdrRecent),
		Filter: func(qr *xorm.Session) *xorm.Session {
			return db.Scope(qr, "@cdr-user", obj.Fromid)
		},
		Limit:  100,
		Order:  "id",
		IsDesc: true,
	}
	proc := func(bean any) (int64, error) {
		obj := bean.(*db.CdrRecent)
		fmt.Print(obj.Id, " >> ")
		print(obj)
		return int64(obj.Id), nil
	}
	if count, err = xquery.Recursive(db.Table(obj), opts, proc, 100); err != nil {
		panic(err)
	}
	fmt.Println(count)

	fmt.Println("执行完成。")
}
