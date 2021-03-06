package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// 可用字符串
const (
	UNICODE_LATIN   = "\u0020-\u007e"                   //英文和主要符号
	UNICODE_PUNCT   = "\u2000-\u206f"                   //标点符号
	UNICODE_CHINESE = "\u4e00-\u9fa5"                   //常用两万汉字
	UNICODE_SPACE   = " \t\r\n"                         //空白类
	UNICODE_NUMBER  = "0-9〇一二三四五六七八九十百千万亿零壹贰叁肆伍陆柒捌玖拾佰仟" //数字类
)

// 字符串比较方式
const (
	CMP_STRING_OMIT             = iota // 不比较
	CMP_STRING_CONTAINS                // 包含
	CMP_STRING_STARTSWITH              // 打头
	CMP_STRING_ENDSWITH                // 结尾
	CMP_STRING_IGNORE_SPACES           // 忽略空格
	CMP_STRING_CASE_INSENSITIVE        // 不分大小写
	CMP_STRING_EQUAL                   // 相等
)

type ConvAction func(s string) string

// 找出其中的数字，不含负号和小数点
func GetNumber(data string) int64 {
	re := regexp.MustCompile("[0-9]+")
	data = re.FindString(data)
	num, err := strconv.ParseInt(data, 10, 64)
	if err == nil {
		return num
	}
	return -1
}

// 分拆为多个部分，并对每一段作处理
func SplitPieces(text, sep string, conv ConvAction) []string {
	pieces := strings.SplitN(text, sep, -1)
	if conv != nil {
		for i, p := range pieces {
			pieces[i] = conv(p)
		}
	}
	return pieces
}

// 删除所有空白，包括中间的
func RemoveSpaces(s string) string {
	subs := map[string]string{
		" ": "", "\n": "", "\r": "", "\t": "", "\v": "", "\f": "",
	}
	return ReplaceWith(s, subs)
}

// 将多个连续空白缩减为一个空格
func ReduceSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// TruncateText 截断长文本
func TruncateText(s string, size int) string {
	if s == "" {
		return ""
	}
	text := []rune(s) // 防止将中文从中间截断
	if size > 0 && len(text) > size {
		s = string(text[:size-3]) + "..."
	}
	return ReduceSpaces(s)
}

// 用:号连接两个部分，如果后一部分也存在的话
func ConcatWith(master, slave string) string {
	if slave != "" {
		master += ":" + slave
	}
	return master
}

// 如果本身不为空，在左右两边添加字符
func WrapWith(s, left, right string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", left, s, right)
}

// 一一对应进行替换，次序不定（因为map的关系）
func ReplaceWith(s string, subs map[string]string) string {
	if s == "" {
		return ""
	}
	var marks []string
	for key, value := range subs {
		marks = append(marks, key, value)
	}
	replacer := strings.NewReplacer(marks...)
	return replacer.Replace(s)
}

// 将方括号替换为反引号，一般用于SQL语句中，因为反引号在Go中是多行文本引号
func ReplaceQuotes(s string) string {
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer("[", "`", "]", "`")
	return replacer.Replace(s)
}

// 比较是否相符
func StringMatch(a, b string, cmp int) bool {
	switch cmp {
	case CMP_STRING_OMIT:
		return true
	case CMP_STRING_CONTAINS:
		return strings.Contains(a, b)
	case CMP_STRING_STARTSWITH:
		return strings.HasPrefix(a, b)
	case CMP_STRING_ENDSWITH:
		return strings.HasSuffix(a, b)
	case CMP_STRING_IGNORE_SPACES:
		a, b = RemoveSpaces(a), RemoveSpaces(b)
		return strings.EqualFold(a, b)
	case CMP_STRING_CASE_INSENSITIVE:
		return strings.EqualFold(a, b)
	default: // 包括 CMP_STRING_EQUAL
		return strings.Compare(a, b) == 0
	}
}

// 是否在字符串列表中，只适用于CMP_STRING_EQUAL和CMP_STRING_STARTSWITH
func compareStringList(x string, lst []string, cmp int) bool {
	size := len(lst)
	if size == 0 {
		return false
	}
	if !sort.StringsAreSorted(lst) {
		sort.Strings(lst)
	}
	i := sort.Search(size, func(i int) bool { return lst[i] >= x })
	return i < size && StringMatch(x, lst[i], cmp)
}

// 是否在字符串列表中
func InStringList(x string, lst []string) bool {
	return compareStringList(x, lst, CMP_STRING_EQUAL)
}

// 是否在字符串列表中，比较方式是有任何一个开头符合
func StartStringList(x string, lst []string) bool {
	return compareStringList(x, lst, CMP_STRING_STARTSWITH)
}

// lst1 是否 lst2 的（真）子集
func IsSubsetList(lst1, lst2 []string, strict bool) bool {
	if len(lst1) > len(lst2) {
		return false
	}
	if strict && len(lst1) == len(lst2) {
		return false
	}
	for _, x := range lst1 {
		if !InStringList(x, lst2) {
			return false
		}
	}
	return true
}

// StrToList 将字符串数组转为一般数组
func StrToList(data []string) []any {
	result := make([]any, len(data))
	for i, v := range data {
		result[i] = v
	}
	return result
}

// MapStrList 将函数应用到每个元素
func MapStrList(data []string, mf func(string) string, ff func(string) bool) []string {
	var result []string
	for _, v := range data {
		if ff != nil && ff(v) == false {
			continue
		}
		if mf != nil {
			v = mf(v)
		}
		result = append(result, v)
	}
	return result
}
