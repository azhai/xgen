package xquery

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func GetIndirectType(v any) (rt reflect.Type) {
	var ok bool
	if rt, ok = v.(reflect.Type); !ok {
		rt = reflect.TypeOf(v)
	}
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return
}

func GetColumns(v any, alias string, cols []string) []string {
	rt := GetIndirectType(v)
	if rt.Kind() != reflect.Struct {
		return cols
	}
	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i).Tag.Get("json")
		if t == "" || t == "-" {
			continue
		} else if strings.HasSuffix(t, "inline") {
			cols = GetColumns(rt.Field(i).Type, alias, cols)
		} else {
			if alias != "" {
				t = fmt.Sprintf("%s.%s", alias, t)
			}
			cols = append(cols, t)
		}
	}
	return cols
}

// QuoteColumns 盲转义，认为字段名以小写字母开头
func QuoteColumns(cols []string, sep string, quote func(string) string) string {
	re := regexp.MustCompile("([a-z][a-zA-Z0-9_]+)")
	repl, origin := quote("$1"), strings.Join(cols, sep)
	result := re.ReplaceAllString(origin, repl)
	if pad := (len(repl) - len("$1")) / 2; pad > 0 {
		left, right := repl[:pad], repl[len(repl)-pad:]
		oldnew := []string{
			left + left, left, right + right, right,
			"'" + left, "'", left + "'", "'",
		}
		result = strings.NewReplacer(oldnew...).Replace(result)
	}
	return result
}
