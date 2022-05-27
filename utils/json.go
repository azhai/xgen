package utils

import (
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
