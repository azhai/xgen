package templater

import (
	"fmt"
)

// 预设的Model相关模板
var (
	golangInitTemplate = `package {{.PkgName}}

import (
	"github.com/azhai/xgen/cmd"
	"github.com/azhai/xgen/dialect"
)


var connCfgs = make(map[string]dialect.ConnConfig)

func PrepareConns(root *config.RootConfig) {
	settings, err := cmd.GetDbSettings(root)
	if err != nil {
		panic(err)
	}
	for _, c := range settings.GetConns() {
		connCfgs[c.Key] = c
	}
}

func GetConnConfig(key string) dialect.ConnConfig {
	if cfg, ok := connCfgs[key]; ok {
		return cfg
	}
	return dialect.ConnConfig{}
}
`

	/**********************************************************************/

	golangModelTemplate = fmt.Sprintf(`package {{.PkgName}}

{{$ilen := len .Imports}}{{if gt $ilen 0 -}}
import ({{- range $imp, $al := .Imports}}
	{{$al}} "{{$imp}}"{{end}}
)
{{end -}}

{{range $class, $table := .Tables}}

// {{$class}} {{$table.Comment}}
type {{$class}} struct { {{- range $table.ColumnsSeq}}{{$col := $table.GetColumn .}}
	{{$col.FieldName}} {{Type $col}} %s{{Tag $table $col "json" "form"}}%s{{end}}
}

// TableName {{$class}}的表名
func (*{{$class}}) TableName() string {
	return "{{$table.Name}}"
}
{{if ne $table.Comment ""}}

// TableComment {{$class}}的备注
func (*{{$class}}) TableComment() string {
	return "{{$table.Comment}}"
}
{{end -}}
{{end}}
`, "`", "`")

	/**********************************************************************/

	golangQueryTemplate = `{{if not .MultipleFiles}}package {{.PkgName}}

import (
	{{- range $imp, $al := .Imports}}{{$al}} "{{$imp}}"
	{{end}}
	xq "github.com/azhai/xgen/xquery"
	"xorm.io/xorm"
)
{{end -}}

{{range $class, $table := .Tables}}
{{$pkey := GetSinglePKey $table -}}
{{$created := GetCreatedColumn $table -}}

// Load the queries of {{$class}}
func (m *{{$class}}) Load(opts ...xq.QueryOption) (bool, error) {
	opts = append(opts, xq.WithTable(m))
	return Query(opts...).Get(m)
}

{{if ne $pkey "" -}}
// Save the queries of {{$class}}
func (m *{{$class}}) Save(changes map[string]any) error {
	return xq.ExecTx(Engine(), func(tx *xorm.Session) (int64, error) {
		if len(changes) == 0 {
			return tx.Table(m).Insert(m)
		} else if m.{{$pkey}} == 0 {
			{{if ne $created "" -}}changes["{{$created}}"] = time.Now()
			{{else}}{{end -}}
			return tx.Table(m).Insert(changes)
		} else {
			return tx.Table(m).ID(m.{{$pkey}}).Update(changes)
		}
	})
}
{{end}}
{{end -}}
`

	/**********************************************************************/

	golangXormTemplate = `package {{.PkgName}}

import (
	{{.AliasName}} "{{.NameSpace}}"
	"github.com/azhai/xgen/dialect"
	xq "github.com/azhai/xgen/xquery"
	_ "{{.Import}}"
	"xorm.io/xorm"
)

var engine *xorm.Engine

// ConnectXorm 连接数据库
func ConnectXorm(cfg dialect.ConnConfig) *xorm.Engine {
	if d := cfg.LoadDialect(); d == nil || !d.IsXormDriver() {
		return nil
	}
	return cfg.QuickConnect(true, true)
}

// Engine 获取当前数据库连接
func Engine() *xorm.Engine {
	if engine == nil {
		cfg := models.GetConnConfig("{{.ConnName}}")
		engine = ConnectXorm(cfg)
		_ = SyncModels(engine)
	}
	return engine
}

// Query 生成查询
func Query(opts ...xq.QueryOption) *xorm.Session {
	qr := Engine().NewSession()
	if len(opts) > 0 {
		return xq.ApplyOptions(qr, opts)
	}
	return qr
}

// Quote 转义表名或字段名
func Quote(value string) string {
	return Engine().Quote(value)
}

// InsertBatch 写入多行数据
func InsertBatch(tableName string, rows ...any) error {
	if len(rows) == 0 {
		return nil
	}
	modify := func(tx *xorm.Session) (int64, error) {
		return tx.Table(tableName).Insert(rows)
	}
	return xq.ExecTx(Engine(), modify)
}

// UpdateBatch 更新多行数据
func UpdateBatch(tableName, pkey string, ids any, changes map[string]any) error {
	if len(changes) == 0 || ids == nil {
		return nil
	}
	modify := func(tx *xorm.Session) (int64, error) {
		return tx.Table(tableName).In(pkey, ids).Update(changes)
	}
	return xq.ExecTx(Engine(), modify)
}

// SyncModels 同步数据库表结构
func SyncModels(eng *xorm.Engine) error {
	if eng == nil {
		return fmt.Errorf("the connection is lost")
	}
	return eng.Sync({{range .Classes}}
		&{{.}}{},{{end}}
	)
}
`

	/**********************************************************************/

	golangRedisTemplate = `package {{.PkgName}}

import (
	{{.AliasName}} "{{.NameSpace}}"

	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/redisw"
)

const (
	SESS_RESCUE_TIMEOUT = 3600                    // 过期前1个小时，重设会话生命期为5个小时
	SESS_CREATE_TIMEOUT = SESS_RESCUE_TIMEOUT * 5 // 最后一次请求后4到5小时会话过期
)

var (
	redisPool *redisw.RedisWrapper
	sessReg  *redisw.SessionRegistry
)

// ConnectRedis 连接数据库
func ConnectRedis(cfg dialect.ConnConfig, db int) *redisw.RedisWrapper {
	if cfg.Type != "redis" {
		return nil
	}
	conn, err := redisw.NewRedisConnDb(cfg, db)
	if err != nil {
		panic(err)
	}
	return redisw.NewRedisConnMux(conn, nil)
}

// Pool 获得连接池
func Pool() *redisw.RedisWrapper {
	if redisPool == nil {
		cfg := models.GetConnConfig("{{.ConnName}}")
		redisPool = ConnectRedis(cfg, -1)
	}
	return redisPool
}

// Registry 获得当前会话管理器
func Registry() *redisw.SessionRegistry {
	if sessReg == nil {
		sessReg = redisw.NewRegistry(Pool())
	}
	return sessReg
}

// Session 获得用户会话
func Session(token string) *redisw.Session {
	sess := Registry().GetSession(token, SESS_CREATE_TIMEOUT)
	timeout := sess.GetTimeout(false)
	if timeout >= 0 && timeout < SESS_RESCUE_TIMEOUT {
		sess.Expire(SESS_CREATE_TIMEOUT) // 重设会话生命期
	}
	return sess
}

// DelSession 删除会话
func DelSession(token string) bool {
	return Registry().DelSession(token)
}
`

	/**********************************************************************/

	golangFlashdbTemplate = `package {{.PkgName}}

import (
	{{.AliasName}} "{{.NameSpace}}"

	"github.com/azhai/xgen/dialect"
	"github.com/arriqaaq/flashdb"
)

var (
	flashConn *flashdb.FlashDB
)

// ConnectFlashDB 连接数据库
func ConnectFlashDB(cfg dialect.ConnConfig) *flashdb.FlashDB {
	if cfg.Type != "flashdb" {
		return nil
	}
	dia := cfg.LoadDialect().(*dialect.FlashDB)
	config := &flashdb.Config{
		Path: dia.Path, EvictionInterval: dia.EvictionInterval,
	}
	conn, err := flashdb.New(config)
	if err != nil {
		panic(err)
	}
	return conn
}

// Singleton 获得连接单例
func Singleton() *flashdb.FlashDB {
	if flashConn == nil {
		cfg := models.GetConnConfig("{{.ConnName}}")
		flashConn = ConnectFlashDB(cfg)
	}
	return flashConn
}
`
)
