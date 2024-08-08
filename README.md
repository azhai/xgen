# xgen

#### 介绍
根据已有数据表生成 models（xorm）代码

如果遇到 UNSIGNED FLOAT 这样的错误，请使用 patch
```bash
go mod tidy
go mod vendor #下载依赖库到vendor目录
cd vendor/xorm.io/xorm/
#xorm 1.3.x
patch -p0 < ../../../xorm-unsigned-double-type.patch
patch -p0 < ../../../xorm-unsigned-double-mysql.patch
#xorm 1.2.x
patch -p1 < ../../../xorm-mysql-unsigned-float.patch
cd -
```

#### 使用
```bash
go mod tidy
#简单使用范例，要求golang 1.18 以上版本
cp settings.hcl.example settings.hcl
make && ./bin/xg
```