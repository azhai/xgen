package main

import (
	"fmt"

	"github.com/azhai/xgen/models"
	db "github.com/azhai/xgen/models/default"
	"github.com/k0kubun/pp"
	"xorm.io/xorm"
)

func runTheDemo() {
	models.Setup()
	db.AddScope("@last", func(qr *xorm.Session, args ...any) *xorm.Session {
		if len(args) > 0 {
			qr = qr.Table(args[0])
		}
		return qr.Desc("id")
	})

	obj := new(db.UserProfileCore)
	query := db.Scope(nil, "@last", obj)
	if _, err := query.Get(obj); err != nil {
		panic(err)
	}
	pp.Println(obj)
	fmt.Println("执行完成。")
}
