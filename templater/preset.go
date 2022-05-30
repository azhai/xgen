package templater

import (
	"fmt"
)

// 预设的Model相关模板
var (
	golangInitTemplate = `package {{.PkgName}}

import (
	"net/url"

	"github.com/azhai/xgen/config"
	"github.com/azhai/xgen/dialect"
)

var (
	connCfgs = make(map[string]dialect.ConnConfig)
	connKeys = url.Values{}
)

func init() {
	if config.IsRunTest() {
		config.BackToDir(1) // 从tests退回根目录
		Setup()
	}
}

func Setup() {
	settings, err := config.ReadConfigFile(nil)
	if err != nil {
		panic(err)
	}
	for _, c := range settings.Conns {
		connCfgs[c.Key] = c
		connKeys.Add(c.Type, c.Key)
	}
}

func GetConnTypes() []string {
	var result []string
	for ct := range connKeys {
		result = append(result, ct)
	}
	return result
}

func GetConnKeys(connType string) []string {
	if keys, ok := connKeys[connType]; ok {
		return keys
	}
	return nil
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
import (
	"database/sql"
	{{range $imp, $al := .Imports}}{{$al}} "{{$imp}}"{{end}}
)
{{end -}}

{{range $table_name, $table := .Tables}}
{{$class := TableMapper $table.Name}}
// ------------------------------------------------------------
// {{$class}} {{$table.Comment}}
// ------------------------------------------------------------
type {{$class}} struct { {{- range $table.ColumnsSeq}}{{$col := $table.GetColumn .}}
	{{ColumnMapper $col.Name}} {{Type $col}} %s{{Tag $table $col true}}%s{{end}}
}

func ({{$class}}) TableName() string {
	return "{{$table_name}}"
}
{{end}}
`, "`", "`")

	/**********************************************************************/

	golangQueryTemplate = `{{if not .MultipleFiles}}package {{.PkgName}}

import (
	{{range $imp, $al := .Imports}}{{$al}} "{{$imp}}"{{end}}
	xq "github.com/azhai/xgen/xquery"
	"xorm.io/xorm"
)
{{end -}}

{{range .Tables}}
{{$class := TableMapper .Name -}}
{{$pkey := GetSinglePKey . -}}
{{$created := GetCreatedColumn . -}}
// ------------------------------------------------------------
// the queries of {{$class}}
// ------------------------------------------------------------

func (m *{{$class}}) Load(opts ...xq.QueryOption) (bool, error) {
	opts = append(opts, xq.WithTable(m))
	return Query(opts...).Get(m)
}

{{if ne $pkey "" -}}
func (m *{{$class}}) Save(changes map[string]any) error {
	return xq.ExecTx(Engine(), func(tx *xorm.Session) (int64, error) {
		if changes == nil || m.{{$pkey}} == 0 {
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

var (
	engine *xorm.Engine
)

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
	}
	return engine
}

// Query 生成查询
func Query(opts ...xq.QueryOption) *xorm.Session {
	qr := Engine().NewSession()
	return xq.ApplyOptions(qr, opts)
}

// Quote 转义表名或字段名
func Quote(value string) string {
	return Engine().Quote(value)
}

// InsertBatch 写入多行数据
func InsertBatch(tableName string, rows []map[string]any) error {
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
