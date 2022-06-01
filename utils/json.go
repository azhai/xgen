package utils

import (
	"database/sql"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// DeepCopy 深度复制对象
func DeepCopy(dest, src any) (err error) {
	var body []byte
	body, err = json.Marshal(src)
	err = json.Unmarshal(body, dest)
	return
}

// PrintJson 以JSON格式输出数据
func PrintJson(data any) (err error) {
	var body []byte
	if body, err = json.Marshal(data); err == nil {
		fmt.Println(string(body))
	}
	return
}

// JsonInt 将被json解码为float64转回整数
func JsonInt64(val any) int64 {
	return int64(val.(float64))
}

// Obj2Dict 将对象转为map格式
func Obj2Dict(obj any) (map[string]any, error) {
	dict := map[string]any{}
	body, err := json.Marshal(obj)
	if err == nil {
		err = json.Unmarshal(body, &dict)
	}
	return dict, err
}

// MarshalJSON json编码
func MarshalJSON[T any](value T, valid bool) ([]byte, error) {
	if valid {
		return json.Marshal(value)
	} else {
		return json.Marshal(nil)
	}
}

// UnmarshalJSON json解码
func UnmarshalJSON[T any](data []byte, value *T) (bool, error) {
	if err := json.Unmarshal(data, value); err != nil {
		return false, err
	}
	return value != nil, nil
}

// NullInt64 可为空整数
type NullInt64 struct {
	sql.NullInt64
}

func (v NullInt64) MarshalJSON() ([]byte, error) {
	return MarshalJSON(v.Int64, v.Valid)
}

func (v *NullInt64) UnmarshalJSON(data []byte) (err error) {
	v.Valid, err = UnmarshalJSON(data, &v.Int64)
	return
}

// NullFloat64 可空浮点数
type NullFloat64 struct {
	sql.NullFloat64
}

func (v NullFloat64) MarshalJSON() ([]byte, error) {
	return MarshalJSON(v.Float64, v.Valid)
}

func (v *NullFloat64) UnmarshalJSON(data []byte) (err error) {
	v.Valid, err = UnmarshalJSON(data, &v.Float64)
	return
}

// // NullFloat64 可空字符串
type NullString struct {
	sql.NullString
}

func (v NullString) MarshalJSON() ([]byte, error) {
	return MarshalJSON(v.String, v.Valid)
}

func (v *NullString) UnmarshalJSON(data []byte) (err error) {
	v.Valid, err = UnmarshalJSON(data, &v.String)
	return
}

// NullTime 可为空时间
type NullTime struct {
	sql.NullTime
}

func (v NullTime) MarshalJSON() ([]byte, error) {
	return MarshalJSON(v.Time, v.Valid)
}

func (v *NullTime) UnmarshalJSON(data []byte) (err error) {
	v.Valid, err = UnmarshalJSON(data, &v.Time)
	return
}
