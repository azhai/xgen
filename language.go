// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"errors"
	"fmt"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/azhai/gozzo/filesystem"
	"github.com/azhai/gozzo/match"
	"github.com/azhai/xgen/rewrite"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"
)

type (
	Formatter func(filename string, sourceCode []byte) ([]byte, error)
	Importter func(tables map[string]*schemas.Table) map[string]string
	Packager  func(targetDir string) string

	kind int
)

const ( // 约定大于配置
	FixedStrMaxSize = 255 // 固定字符串最大长度

	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	integerKind
	stringKind
	uintKind

	XormTagName       = "xorm"
	XormTagNotNull    = "notnull"
	XormTagAutoIncr   = "autoincr"
	XormTagPrimaryKey = "pk"
	XormTagUnique     = "unique"
	XormTagIndex      = "index"
)

var (
	errBadComparisonType = errors.New("invalid type for comparison")
	// errBadComparison     = errors.New("incompatible types for comparison")
	// errNoComparison      = errors.New("missing argument for comparison")

	TypeOfString  = reflect.TypeOf("")
	TypeOfBytes   = reflect.TypeOf([]byte{})
	TypeOfBool    = reflect.TypeOf(false)
	TypeOfInt     = reflect.TypeOf(0)
	TypeOfInt64   = reflect.TypeOf(int64(0))
	TypeOfFloat64 = reflect.TypeOf(float64(0))
	TypeOfTime    = reflect.TypeOf(time.Time{})

	golang = &Language{ // Golang represents a golang language
		Name:     "golang",
		ExtName:  ".go",
		Template: nil,
		Types:    map[string]string{},
		Funcs: template.FuncMap{
			"Type":             type2string,
			"Tag":              tag2string,
			"GetSinglePKey":    getSinglePKey,
			"GetCreatedColumn": getCreatedColumn,
		},
		Formatter: rewrite.WriteGolangFilePrettify,
		Importter: genGoImports,
		Packager:  genNameSpace,
	}
)

// Language represents a languages supported when reverse codes
type Language struct {
	Name      string
	ExtName   string
	Template  *template.Template
	Types     map[string]string
	Funcs     template.FuncMap
	Formatter Formatter
	Importter Importter
	Packager  Packager
}

func escapeTag(value string) string {
	return strings.ReplaceAll(value, "#", "")
}

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}
	return invalidKind, errBadComparisonType
}

// sqlType2Type default sql type change to go types
func sqlType2Type(st schemas.SQLType) (rtype reflect.Type, rtstr string) {
	name := strings.ToUpper(st.Name)
	rtype = TypeOfString
	switch name {
	case schemas.Bool:
		rtype = TypeOfBool
	case schemas.Bit, schemas.UnsignedBit, schemas.TinyInt, schemas.UnsignedTinyInt:
		if st.DefaultLength == 1 {
			rtype = TypeOfBool
		} else {
			rtype = TypeOfInt
		}
	case schemas.SmallInt, schemas.MediumInt, schemas.Int, schemas.Integer, schemas.Serial,
		schemas.UnsignedSmallInt, schemas.UnsignedMediumInt, schemas.UnsignedInt:
		rtype = TypeOfInt
	case schemas.BigInt, schemas.BigSerial, schemas.UnsignedBigInt:
		rtype = TypeOfInt64
	case schemas.Float, schemas.Real, schemas.Double:
		rtype = TypeOfFloat64
	case schemas.DateTime, schemas.Date, schemas.Time, schemas.TimeStamp,
		schemas.TimeStampz, schemas.SmallDateTime, schemas.Year:
		rtype = TypeOfTime
	case schemas.TinyBlob, schemas.Blob, schemas.MediumBlob, schemas.LongBlob,
		schemas.Bytea, schemas.Binary, schemas.VarBinary, schemas.UniqueIdentifier:
		rtype, rtstr = TypeOfBytes, "[]byte"
	case schemas.Varchar, schemas.NVarchar:
		if st.DefaultLength > FixedStrMaxSize {
			rtstr = "xutils.NullString"
		}
	case schemas.TinyText, schemas.Text,
		schemas.NText, schemas.MediumText, schemas.LongText:
		rtstr = "xutils.NullString"
		// case schemas.Char, schemas.NChar, schemas.Enum, schemas.Set, schemas.Uuid, schemas.Clob, schemas.SysName:
		//	rtstr = rtype.String()2
		// case schemas.Decimal, schemas.Numeric, schemas.Money, schemas.SmallMoney:
		//	rtstr = rtype.String()
	}
	if rtstr == "" {
		rtstr = rtype.String()
	}
	return
}

func getCol(cols map[string]*schemas.Column, name string) *schemas.Column {
	return cols[strings.ToLower(name)]
}

func getSinglePKey(table *schemas.Table) string {
	if cols := table.PKColumns(); len(cols) == 1 {
		return cols[0].FieldName
	}
	return ""
}

func getCreatedColumn(table *schemas.Table) string {
	for name, ok := range table.Created {
		if ok {
			return table.GetColumn(name).Name
		}
	}
	if col := table.GetColumn("created_at"); col != nil {
		if col.SQLType.IsTime() {
			return "created_at"
		}
	}
	return ""
}

func trimAnyPrefix(word string, prefixes []string) string {
	if len(prefixes) == 0 {
		return word
	}
	size := len(word)
	for _, pre := range prefixes {
		word = strings.TrimPrefix(word, pre)
		if len(word) < size { // 成功
			return word
		}
	}
	return word
}

func convertMapper(mapname string) names.Mapper {
	switch mapname {
	case "gonic":
		return names.LintGonicMapper
	case "same":
		return names.SameMapper{}
	default:
		return names.SnakeMapper{}
	}
}

func genNameSpace(targetDir string) string {
	// 先重试提取已有代码文件（排除测试代码）的包名
	files, err := filesystem.FindFiles(targetDir, ".go")
	if err == nil && len(files) > 0 {
		for filename := range files {
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}
			cp, err := rewrite.NewFileParser(filename)
			if err != nil {
				continue
			}
			if nameSpace := cp.GetPackage(); nameSpace != "" {
				return nameSpace
			}
		}
	}
	// 否则直接使用目录名，需要排除Golang关键词
	nameSpace := strings.ToLower(filepath.Base(targetDir))
	return aliasNameSpace(nameSpace)
}

func aliasNameSpace(nameSpace string) string {
	if nameSpace == "default" { // it is golang keyword
		nameSpace = "db"
	} else if token.IsKeyword(nameSpace) {
		nameSpace = "db" + nameSpace
	}
	return nameSpace
}

func genGoImports(tables map[string]*schemas.Table) map[string]string {
	imports := make(map[string]string)
	for _, table := range tables {
		for _, col := range table.Columns() {
			s := type2string(col)
			if s == "time.Time" || s == "xutils.NullTime" {
				imports["time"] = ""
			}
			if strings.HasPrefix(s, "xutils.Null") {
				imports["github.com/azhai/xgen/utils"] = "xutils"
			}
		}
	}
	return imports
}

func type2string(col *schemas.Column) string {
	t, s := sqlType2Type(col.SQLType)
	if s == "string" && col.Nullable {
		s = "xutils.NullString"
	}
	if t != TypeOfBool {
		return s
	}
	if strings.HasPrefix(col.Name, "is_") ||
		strings.HasPrefix(col.Name, "not_") {
		return s
	}
	return TypeOfInt.String()
}

func tag2string(table *schemas.Table, col *schemas.Column, names ...string) string {
	tx := tagXorm(table, col)
	if len(names) == 0 {
		return tx
	}
	var ts []string
	for _, name := range names {
		ts = append(ts, tagCustom(col, name))
	}
	if tx != "" {
		ts = append(ts, tx)
	}
	return strings.Join(ts, " ")
}

func tagCustom(col *schemas.Column, name string) string {
	if col.Name == "" {
		return ""
	}
	return fmt.Sprintf(`%s:"%s"`, name, col.Name)
}

func tagXorm(table *schemas.Table, col *schemas.Column) string {
	isNameId := col.FieldName == "Id"
	isIdPk := isNameId && type2string(col) == "int64"

	var res []string
	if !col.Nullable {
		if !isIdPk {
			res = append(res, XormTagNotNull)
		}
	}
	if col.IsPrimaryKey {
		res = append(res, XormTagPrimaryKey)
	}
	if col.Default != "" {
		res = append(res, "default "+col.Default)
	}
	if col.IsAutoIncrement {
		res = append(res, XormTagAutoIncr)
	}

	if col.SQLType.IsTime() {
		lowerName := strings.ToLower(col.Name)
		if strings.HasPrefix(lowerName, "created") {
			res = append(res, "created")
		} else if strings.HasPrefix(lowerName, "updated") {
			res = append(res, "updated")
		} else if strings.HasPrefix(lowerName, "deleted") {
			res = append(res, "deleted")
		}
	}

	if comm := match.TruncateText(col.Comment, 50); comm != "" {
		comm = template.HTMLEscapeString(comm) // 备注脱敏
		res = append(res, fmt.Sprintf("comment('%s')", comm))
	}

	names := make([]string, 0, len(col.Indexes))
	for name := range col.Indexes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		index := table.Indexes[name]
		var uistr string
		if index.Type == schemas.UniqueType {
			uistr = XormTagUnique
		} else if index.Type == schemas.IndexType {
			uistr = XormTagIndex
		}
		if len(index.Cols) > 1 {
			uistr += "(" + index.Name + ")"
		}
		res = append(res, uistr)
	}

	res = append(res, GetColTypeString(col))
	if len(res) > 0 {
		tagValue := escapeTag(strings.Join(res, " ")) // 脱敏处理
		return fmt.Sprintf(`%s:"%s"`, XormTagName, tagValue)
	}
	return ""
}

// GetColTypeString get the col type include length, for example: VARCHAR(255)
func GetColTypeString(col *schemas.Column) string {
	ctstr := col.SQLType.Name
	if col.Length != 0 {
		if col.Length2 != 0 { // float, decimal
			ctstr += fmt.Sprintf("(%v,%v)", col.Length, col.Length2)
		} else { // int, char, varchar
			ctstr += fmt.Sprintf("(%v)", col.Length)
		}
	} else if len(col.EnumOptions) > 0 { // enum
		ctstr += "("
		opts := ""

		enumOptions := make([]string, 0, len(col.EnumOptions))
		for enumOption := range col.EnumOptions {
			enumOptions = append(enumOptions, enumOption)
		}
		sort.Strings(enumOptions)

		for _, v := range enumOptions {
			opts += fmt.Sprintf(",'%v'", v)
		}
		ctstr += strings.TrimLeft(opts, ",")
		ctstr += ")"
	} else if len(col.SetOptions) > 0 { // set
		ctstr += "("
		opts := ""

		setOptions := make([]string, 0, len(col.SetOptions))
		for setOption := range col.SetOptions {
			setOptions = append(setOptions, setOption)
		}
		sort.Strings(setOptions)

		for _, v := range setOptions {
			opts += fmt.Sprintf(",'%v'", v)
		}
		ctstr += strings.TrimLeft(opts, ",")
		ctstr += ")"
	}
	return ctstr
}
