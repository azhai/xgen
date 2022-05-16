package redisw

import (
	"fmt"
	"time"

	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/utils"

	"github.com/gomodule/redigo/redis"
	"github.com/gomodule/redigo/redisx"
)

const (
	REDIS_DEFAULT_IDLE_CONN    = 3   // 最大空闲连接数
	REDIS_DEFAULT_IDLE_TIMEOUT = 240 // 最大空闲时长，单位：秒
	REDIS_DEFAULT_EXEC_RETRY   = 3   // 重试次数
	REDIS_DEFAULT_READ_TIMEOUT = 7   // 命令最大执行时长，单位：秒
)

var (
	StrToList      = utils.StrToList // 将字符串数组转为一般数组
	KeysEmptyError = fmt.Errorf("the param which named 'keys' must not empty !")
)

// redigo没有将应答中的OK转为bool值(2020-01-16)
func ReplyBool(reply any, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	var answer string
	answer, err = redis.String(reply, err)
	return answer == "OK", err
}

// RedisContainer Redis容器，包括 *redis.Pool 和 *redisx.ConnMux 两个实现
type RedisContainer interface {
	Get() redis.Conn
	Close() error
}

// RedisWrapper Redis包装器，给容器加上超时等参数
type RedisWrapper struct {
	MaxIdleConn int // 最大空闲连接数
	MaxIdleTime int // 最大空闲时长
	RetryTimes  int // 重试次数
	MaxReadTime int // 命令最大执行时长（不算连接部分）
	RedisContainer
}

// NewRedisWrapper 包装一个Redis空容器
func NewRedisWrapper() *RedisWrapper {
	return &RedisWrapper{
		MaxIdleConn: REDIS_DEFAULT_IDLE_CONN,
		MaxIdleTime: REDIS_DEFAULT_IDLE_TIMEOUT,
		RetryTimes:  REDIS_DEFAULT_EXEC_RETRY,
		MaxReadTime: REDIS_DEFAULT_READ_TIMEOUT,
	}
}

// NewRedisConn 根据配置连接redis
// cfg = dialect.ConnConfig{
// 	DSN: (dialect.Redis{
// 		Host:     "127.0.0.1",
// 		Database: 3,
// 	}).BuildDSN(),
// 	Password: "secrect",
// }
func NewRedisConn(cfg dialect.ConnConfig) (redis.Conn, error) {
	return NewRedisConnDb(cfg, -1)
}

// NewRedisConnDb 建立Redis实际连接
func NewRedisConnDb(cfg dialect.ConnConfig, db int) (redis.Conn, error) {
	var opts []redis.DialOption
	if cfg.Username != "" {
		opts = append(opts, redis.DialUsername(cfg.Username))
	}
	if cfg.Password != "" {
		opts = append(opts, redis.DialPassword(cfg.Password))
	}
	if db >= 0 {
		opts = append(opts, redis.DialDatabase(db))
	}
	return redis.DialURL(cfg.GetDSN(false), opts...)
}

// NewRedisConnMux 建立Redis连接复用
func NewRedisConnMux(conn redis.Conn, err error) *RedisWrapper {
	r := NewRedisWrapper()
	r.MaxReadTime = 0 // 不支持 ConnWithTimeout 和 DoWithTimeout
	if err == nil {
		r.RedisContainer = redisx.NewConnMux(conn)
	}
	return r
}

// NewRedisPool 建立Redis连接池
func NewRedisPool(cfg dialect.ConnConfig, maxIdle int) *RedisWrapper {
	r := NewRedisWrapper()
	if maxIdle >= 0 {
		r.MaxIdleConn = maxIdle
	}
	timeout := time.Second * time.Duration(r.MaxIdleTime)
	r.RedisContainer = &redis.Pool{
		MaxIdle: r.MaxIdleConn, IdleTimeout: timeout,
		Dial: func() (redis.Conn, error) {
			return NewRedisConn(cfg)
		},
	}
	return r
}

// GetMaxReadDuration 单命令最大执行时长（不算连接部分）
func (r *RedisWrapper) GetMaxReadDuration() time.Duration {
	if r.MaxReadTime > 0 {
		return time.Second * time.Duration(r.MaxReadTime)
	}
	return 0
}

// Exec 执行命令，将会重试几次
func (r *RedisWrapper) Exec(cmd string, args ...any) (any, error) {
	var (
		err   error
		reply any
	)
	mrd := r.GetMaxReadDuration()
	for i := 0; i < r.RetryTimes; i++ {
		if mrd > 0 {
			reply, err = redis.DoWithTimeout(r.Get(), mrd, cmd, args...)
		} else {
			reply, err = r.Get().Do(cmd, args...)
		}
		if err == nil {
			break
		}
	}
	return reply, err
}

// GetSize 当前db缓存键的数量
func (r *RedisWrapper) GetSize() int {
	size, _ := redis.Int(r.Exec("DBSIZE"))
	return size
}

// GetTimeout -1=无限 -2=不存在 -3=出错
func (r *RedisWrapper) GetTimeout(key string) int {
	sec, err := redis.Int(r.Exec("TTL", key))
	if err == nil {
		return sec
	}
	return -3
}

// Expire 设置键的过期时间
func (r *RedisWrapper) Expire(key string, timeout int) (bool, error) {
	reply, err := r.Exec("EXPIRE", key, timeout)
	return ReplyBool(reply, err)
}

func (r *RedisWrapper) Delete(keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, KeysEmptyError
	}
	reply, err := r.Exec("DEL", StrToList(keys)...)
	return redis.Int(reply, err)
}

// DeleteAll 清空当前db
func (r *RedisWrapper) DeleteAll() (bool, error) {
	return ReplyBool(r.Exec("FLUSHDB"))
}

func (r *RedisWrapper) Exists(key string) (bool, error) {
	return ReplyBool(r.Exec("EXISTS", key))
}

// Find 模糊查找
func (r *RedisWrapper) Find(wildcard string) ([]string, error) {
	return redis.Strings(r.Exec("KEYS", wildcard))
}

func (r *RedisWrapper) Rename(old, dst string) (bool, error) {
	return ReplyBool(r.Exec("RENAME", old, dst))
}
