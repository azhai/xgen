package redisw

import (
	"github.com/gomodule/redigo/redis"
)

type RedisHash struct {
	name    string
	timeout int
	*RedisWrapper
}

func NewRedisHash(r *RedisWrapper, name string, timeout int) *RedisHash {
	return &RedisHash{RedisWrapper: r, name: name, timeout: timeout}
}

func (rh *RedisHash) Exec(cmd string, args ...any) (any, error) {
	args = append([]any{rh.name}, args...)
	return rh.RedisWrapper.Exec(cmd, args...)
}

// GetSize 获取哈希表元素数量
func (rh *RedisHash) GetSize() int {
	size, _ := redis.Int(rh.Exec("HLEN"))
	return size
}

// GetTimeout 获取剩余时间 -1=无限 -2=不存在 -3=出错
func (rh *RedisHash) GetTimeout(predict bool) int {
	timeout := rh.RedisWrapper.GetTimeout(rh.name)
	if timeout == -2 && predict { // 尚未设置，使用预定值
		timeout = rh.timeout
	}
	return timeout
}

func (rh *RedisHash) GetKeys() []string {
	keys, _ := redis.Strings(rh.Exec("HKEYS"))
	return keys
}

func (rh *RedisHash) Expire(timeout int) (bool, error) {
	return rh.RedisWrapper.Expire(rh.name, timeout)
}

func (rh *RedisHash) OrigDelete(keys ...any) (int, error) {
	if len(keys) == 0 {
		return 0, KeysEmptyError
	}
	reply, err := rh.Exec("HDEL", keys...)
	return redis.Int(reply, err)
}

func (rh *RedisHash) Delete(keys ...string) (int, error) {
	return rh.OrigDelete(StrToList(keys)...)
}

func (rh *RedisHash) DeleteAll() (bool, error) {
	affects, err := rh.RedisWrapper.Delete(rh.name)
	return affects > 0, err
}

func (rh *RedisHash) Exists(key string) (bool, error) {
	return ReplyBool(rh.Exec("HEXISTS", key))
}

func (rh *RedisHash) SetNX(key string, value any) (int, error) {
	affects, err := redis.Int(rh.Exec("HSETNX", key, value))
	if affects == 1 {
		rh.Exec("EXPIRE", rh.timeout)
	}
	return affects, err
}

func (rh *RedisHash) SetVal(key string, value any) (int, error) {
	defer rh.Exec("EXPIRE", rh.timeout)
	return redis.Int(rh.Exec("HSET", key, value))
}

func (rh *RedisHash) GetVal(key string) (any, error) {
	return rh.Exec("HGET", key)
}

func (rh *RedisHash) GetString(key string) (string, error) {
	return redis.String(rh.GetVal(key))
}

func (rh *RedisHash) GetInt(key string) (int, error) {
	return redis.Int(rh.GetVal(key))
}

func (rh *RedisHash) IncrInt(key string, offset int) (int, error) {
	value, err := rh.IncrInt64(key, int64(offset))
	return int(value), err
}

func (rh *RedisHash) GetInt64(key string) (int64, error) {
	return redis.Int64(rh.GetVal(key))
}

func (rh *RedisHash) IncrInt64(key string, offset int64) (int64, error) {
	return redis.Int64(rh.Exec("HINCRBY", key, offset))
}

func (rh *RedisHash) GetFloat(key string) (float64, error) {
	return redis.Float64(rh.GetVal(key))
}

func (rh *RedisHash) GetAll() (any, error) {
	return rh.Exec("HGETALL")
}

func (rh *RedisHash) GetAllString() (map[string]string, error) {
	return redis.StringMap(rh.GetAll())
}
