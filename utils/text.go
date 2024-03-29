package utils

import (
	"fmt"
)

// ConcatWith 用:号连接两个部分，如果后一部分也存在的话
func ConcatWith(master, slave string) string {
	if slave != "" {
		master += ":" + slave
	}
	return master
}

// WrapWith 如果本身不为空，在左右两边添加字符
func WrapWith(s, left, right string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", left, s, right)
}

// StrToList 将字符串数组转为一般数组
func StrToList(data []string) []any {
	result := make([]any, len(data))
	for i, v := range data {
		result[i] = v
	}
	return result
}
