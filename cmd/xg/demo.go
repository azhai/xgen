package main

import (
	"fmt"
	// "github.com/azhai/xgen/models"
	// db "github.com/azhai/xgen/models/default"
	// "github.com/k0kubun/pp"
	// "xorm.io/xorm"
)

func runTheDemo() {
	// models.Setup()
	// db.AddScope("@last", func(qr *xorm.Session, args ...any) *xorm.Session {
	// 	if len(args) > 0 {
	// 		qr = qr.Table(args[0])
	// 	}
	// 	return qr.Desc("id")
	// })

	// var err error
	// obj := new(db.Menu)
	// query := db.Scope(nil, "@last", obj)
	// if _, err = query.Get(obj); err != nil {
	// 	panic(err)
	// }
	// pp.Println(obj)

	// var rows []*db.Menu
	// count := int64(0)
	// query = db.Scope(db.Table(obj), "@id-in", []int{2, 6, 8})
	// if count, err = query.FindAndCount(&rows); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("there are N rows, N =", count)
	// pp.Println(rows)

	fmt.Println("执行完成。")
}
