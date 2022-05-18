// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/rewrite"
	"github.com/azhai/xgen/templater"
	"github.com/azhai/xgen/utils"

	"xorm.io/xorm/schemas"
)

const ( // 约定大于配置
	INIT_FILE_NAME   = "init"
	CONN_FILE_NAME   = "conn"
	SINGLE_FILE_NAME = "models"
	QUERY_FILE_NAME  = "queries"
)

// ReverseConfig table反转为model配置
type ReverseConfig struct {
	Language          string   `hcl:"language,label" json:"language,omitempty"`
	OutputDir         string   `hcl:"output_dir" json:"output_dir"`
	NameSpace         string   `hcl:"name_space" json:"name_space"`
	MultipleFiles     bool     `hcl:"multiple_files,optional" json:"multiple_files,omitempty"`
	TablePrefix       string   `hcl:"table_prefix,optional" json:"table_prefix,omitempty"`
	IncludeTables     []string `hcl:"include_tables,optional" json:"include_tables,omitempty"`
	ExcludeTables     []string `hcl:"exclude_tables,optional" json:"exclude_tables,omitempty"`
	TableMapper       string   `hcl:"table_mapper,optional" json:"table_mapper,omitempty"`
	ColumnMapper      string   `hcl:"column_mapper,optional" json:"column_mapper,omitempty"`
	MixinDir          string   `hcl:"mixin_dir,optional" json:"mixin_dir,omitempty"`
	MixinNS           string   `hcl:"mixin_ns,optional" json:"mixin_ns,omitempty"`
	ModelTemplatePath string   `hcl:"model_template_path,optional" json:"model_template_path,omitempty"`
	QueryTemplatePath string   `hcl:"query_template_path,optional" json:"query_template_path,omitempty"`
}

// GetTablePrefixes 获取可用表名前缀
func (c ReverseConfig) GetTablePrefixes() []string {
	prefixes := []string{}    // 不使用表名前缀
	if c.TablePrefix == "*" { // 使用所有的包含列标前缀
		for _, pre := range c.IncludeTables {
			pre = strings.TrimRight(pre, "*")
			if !strings.Contains(pre, "*") { // 原正则式是右匹配
				prefixes = append(prefixes, pre)
			}
		}
	} else if c.TablePrefix != "" { // 仅一个固定前缀
		prefixes = append(prefixes, c.TablePrefix)
	}
	return prefixes
}

// GetTemplateName 获取模板名称，优先使用配置，然后是预设模板
func (c ReverseConfig) GetTemplateName(name string) string {
	switch strings.ToLower(name) {
	default:
		return ""
	case "model":
		if c.ModelTemplatePath != "" {
			return c.ModelTemplatePath
		}
		return "model"
	case "query":
		if c.QueryTemplatePath != "" {
			return c.QueryTemplatePath
		}
		return "query"
	}
}

func (c ReverseConfig) PrepareMixins() (mixinNames []string) {
	if c.MixinDir == "" {
		return
	}
	files, _ := rewrite.FindFiles(c.MixinDir, ".go")
	for filename := range files {
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		newNames := rewrite.AddFormerMixins(filename, c.MixinNS, "")
		mixinNames = append(mixinNames, newNames...)
	}
	return
}

// Reverser model反转器
type Reverser struct {
	currOutDir string
	lang       *Language
	target     ReverseConfig
}

// NewGoReverser 创建Golang反转器
func NewGoReverser(target ReverseConfig) *Reverser {
	target.PrepareMixins()
	return &Reverser{lang: golang, target: target}
}

// GetFormatter 对应语言的美化代码工具
func (r *Reverser) GetFormatter() Formatter {
	if r.lang == nil || r.lang.Formatter == nil {
		return rewrite.WriteCodeFile
	}
	return r.lang.Formatter
}

// SetOutDir 设置输出子目录
func (r *Reverser) SetOutDir(key string) string {
	if key == "" {
		r.currOutDir = r.target.OutputDir
	} else {
		r.currOutDir = filepath.Join(r.target.OutputDir, key)
	}
	os.MkdirAll(r.currOutDir, rewrite.DEFAULT_DIR_MODE)
	return r.currOutDir
}

// GetOutFileName 获取输出文件名
func (r *Reverser) GetOutFileName(name string) string {
	return filepath.Join(r.currOutDir, name+r.lang.ExtName)
}

// GenModelInitFile 生成models目录下的init文件
func (r *Reverser) GenModelInitFile(tmplName string) error {
	r.SetOutDir("")
	pkgName := filepath.Base(r.target.NameSpace)
	if pkgName == "" {
		pkgName = "models"
	}
	data := map[string]any{"PkgName": pkgName}

	tmpl := templater.LoadTemplate(tmplName, nil)
	codeText, err := templater.RenderTemplate(tmpl, data)
	if err == nil {
		filename := r.GetOutFileName(INIT_FILE_NAME)
		formatter := r.GetFormatter()
		_, err = formatter(filename, codeText)
	}
	return err
}

// ExecuteReverse 生成单个数据库下的代码文件，一个数据库一个子目录
func (r *Reverser) ExecuteReverse(source dialect.ConnConfig, interActive, verbose bool) error {
	r.SetOutDir(source.Key)
	dia := source.LoadDialect()
	pkgName := aliasNameSpace(source.Key)
	data := map[string]any{
		"PkgName":   pkgName,
		"ConnName":  source.Key,
		"NameSpace": r.target.NameSpace,
		"AliasName": "models",
		"Import":    dia.ImporterPath(),
	}
	if strings.HasSuffix(r.target.NameSpace, "/models") {
		data["AliasName"] = ""
	}

	tmplName := "xorm"
	if source.Type == "redis" || source.Type == "flashdb" {
		tmplName = source.Type
	} else {
		tableSchemas, err := source.QuickConnect(verbose, verbose).DBMetas()
		if err != nil {
			fmt.Println(err)
		}
		tableSchemas = FilterTables(tableSchemas, r.target.IncludeTables, r.target.ExcludeTables, 4)
		if len(tableSchemas) == 0 {
			return nil
		}
		pkgName := aliasNameSpace(source.Key)
		if err = r.ReverseTables(pkgName, tableSchemas); err != nil {
			return err
		}
	}
	tmpl := templater.LoadTemplate(tmplName, nil)
	codeText, err := templater.RenderTemplate(tmpl, data)
	if err == nil {
		filename := r.GetOutFileName(CONN_FILE_NAME)
		formatter := r.GetFormatter()
		if _, err = formatter(filename, codeText); err == nil {
			err = ApplyDirMixins(r.currOutDir, verbose)
		}
	}
	return err
}

// ReverseTables 生成单个数据的model和query文件，或者一张表一个文件（当MultipleFiles=true）
func (r *Reverser) ReverseTables(pkgName string, tableSchemas []*schemas.Table) error {
	tbMapper := convertMapper(r.target.TableMapper).Table2Obj
	colMapper := convertMapper(r.target.ColumnMapper).Table2Obj
	funcs := r.lang.Funcs
	funcs["TableMapper"], funcs["ColumnMapper"] = tbMapper, colMapper
	tables := make(map[string]*schemas.Table)
	tablePrefixes := r.target.GetTablePrefixes()
	for _, table := range tableSchemas {
		fmt.Println("-", pkgName, table.Name)
		tableName := table.Name
		if len(tablePrefixes) > 0 {
			table.Name = trimAnyPrefix(table.Name, tablePrefixes)
		}
		for _, col := range table.Columns() {
			col.FieldName = colMapper(col.Name)
		}
		tables[tableName] = table
	}
	data := map[string]any{
		"PkgName":       pkgName,
		"NameSpace":     r.target.NameSpace,
		"MultipleFiles": r.target.MultipleFiles,
	}

	formatter := r.GetFormatter()
	importter := r.lang.Importter
	tmpl := r.lang.Template
	if tmpl == nil {
		tmplName := r.target.GetTemplateName("model")
		tmpl = templater.LoadTemplate(tmplName, funcs)
	}
	if r.target.MultipleFiles { // 每张表一个文件
		for tableName, table := range tables {
			tbs := map[string]*schemas.Table{tableName: table}
			data["Tables"], data["Imports"] = tbs, importter(tbs)
			codeText, err := templater.RenderTemplate(tmpl, data)
			if err != nil {
				return err
			}
			filename := r.GetOutFileName(table.Name)
			if _, err = formatter(filename, codeText); err != nil {
				return err
			}
		}
	} else {
		data["Tables"], data["Imports"] = tables, importter(tables)
		codeText, err := templater.RenderTemplate(tmpl, data)
		if err != nil {
			return err
		}
		filename := r.GetOutFileName(SINGLE_FILE_NAME)
		if _, err = formatter(filename, codeText); err != nil {
			return err
		}

		tmplName := r.target.GetTemplateName("query")
		tmpl = templater.LoadTemplate(tmplName, funcs)
		data["Imports"] = map[string]string{"xorm.io/xorm": ""}
		codeText, err = templater.RenderTemplate(tmpl, data)
		if err != nil {
			return err
		}
		filename = r.GetOutFileName(QUERY_FILE_NAME)
		if _, err = formatter(filename, codeText); err != nil {
			return err
		}
	}
	return nil
}

// FilterTables 按照ExcludeTables和IncludeTables配置过滤数据表
func FilterTables(tables []*schemas.Table, includes, excludes []string, tailDigits int) []*schemas.Table {
	res := make([]*schemas.Table, 0, len(tables))
	inclMatchers, exclMatchers := utils.NewGlobs(includes), utils.NewGlobs(excludes)
	digitsReg := regexp.MustCompile(fmt.Sprintf("_[0-9]{%d,}", tailDigits))
	for _, tb := range tables {
		// 排除4个数字以上结尾的分表
		if tailDigits > 0 && digitsReg.MatchString(tb.Name) {
			continue
		}
		if exclMatchers.MatchAny(tb.Name, false) {
			continue
		}
		if inclMatchers.MatchAny(tb.Name, true) {
			res = append(res, tb)
		}
	}
	return res
}

// ApplyDirMixins 将已知的Mixin嵌入到匹配的Model中
func ApplyDirMixins(currDir string, verbose bool) error {
	cps := rewrite.NewComposer()
	files, _ := rewrite.FindFiles(currDir, ".go")
	var err error
	for filename := range files {
		_err := rewrite.ParseAndMixinFile(cps, filename, verbose)
		if _err != nil {
			err = _err
		}
	}
	return err
}
