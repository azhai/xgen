package rewrite

import (
	"bytes"
	"fmt"
	"go/ast"
	"sort"
	"strings"

	"github.com/azhai/xgen/utils/enums"
)

const MODEL_EXTENDS = "`json:\",inline\" xorm:\"extends\"`"

var (
	substituteNames  = []string{"xquery.TimeMixin", "xquery.NestedMixin"}
	substituteModels = map[string]*ModelSummary{
		"xquery.TimeMixin": {
			Name:   "xquery.TimeMixin",
			Import: "github.com/azhai/xgen/xquery",
			Alias:  "",
			FieldLines: []string{
				"CreatedAt time.Time `json:\"created_at\" xorm:\"created comment('创建时间') TIMESTAMP\"`       // 创建时间",
				"UpdatedAt time.Time `json:\"updated_at\" xorm:\"updated comment('更新时间') TIMESTAMP\"`       // 更新时间",
				"DeletedAt time.Time `json:\"deleted_at\" xorm:\"deleted comment('删除时间') index TIMESTAMP\"` // 删除时间",
			},
		},
		"xquery.NestedMixin": {
			Name:   "xquery.NestedMixin",
			Import: "github.com/azhai/xgen/xquery",
			Alias:  "",
			FieldLines: []string{
				"Lft   int `json:\"lft\" xorm:\"notnull default 0 comment('左边界') INT(10)\"`           // 左边界",
				"Rgt   int `json:\"rgt\" xorm:\"notnull default 0 comment('右边界') index INT(10)\"`     // 右边界",
				"Depth int `json:\"depth\" xorm:\"notnull default 1 comment('高度') index TINYINT(3)\"` // 高度",
			},
		},
	}
)

// RegisterSubstitute 注册可替换Model
func RegisterSubstitute(sub *ModelSummary) {
	if sub != nil && !sub.IsExists {
		substituteModels[sub.Name] = sub
		substituteNames = append(substituteNames, sub.Name)
	}
}

// RemoveSubstitute 删除可替换Model
func RemoveSubstitute(name string) {
	if _, ok := substituteModels[name]; ok {
		substituteModels[name] = nil
	}
}

// ModelSummary Model摘要
type ModelSummary struct {
	Name           string
	Substitute     string
	Import, Alias  string
	Features       []string
	sortedFeatures []string
	FieldLines     []string
	IsChanged      bool
	IsExists       bool //同构Model已存在
}

// GetInnerCode 找出 model 内部代码，即在 {} 里面的内容
func (s ModelSummary) GetInnerCode() string {
	var buf bytes.Buffer
	for _, line := range s.FieldLines {
		buf.WriteString(fmt.Sprintf("\t%s\n", line))
	}
	return buf.String()
}

// GetSortedFeatures 找出 model 的所有特征并排序
func (s ModelSummary) GetSortedFeatures() []string {
	if len(s.sortedFeatures) > 0 {
		return s.sortedFeatures
	}
	size := len(s.FieldLines)
	if len(s.Features) != size {
		s.Features = make([]string, size)
		for i, line := range s.FieldLines {
			s.Features[i] = GetLineFeature(line)
		}
	}
	s.sortedFeatures = append([]string{}, s.Features...)
	sort.Strings(s.sortedFeatures)
	return s.sortedFeatures
}

// Isomorph 已经是
func (s *ModelSummary) Isomorph() bool {
	features := s.GetSortedFeatures()
	return len(features) == 1 && strings.HasSuffix(features[0], ":inline")
}

// GetSubstitute
func (s *ModelSummary) GetSubstitute() string {
	if s.Substitute == "" {
		s.Substitute = fmt.Sprintf("*%s %s", s.Name, MODEL_EXTENDS)
	}
	return s.Substitute
}

// ParseFields 解析 struct 代码，提取特征并补充注释到代码
func (s *ModelSummary) ParseFields(cp *CodeParser, node *DeclNode) int {
	size := len(node.Fields)
	s.Features = make([]string, size)
	s.FieldLines = make([]string, size)
	for i, f := range node.Fields {
		code := cp.GetNodeCode(f)
		if feat := GetLineFeature(code); feat != "" {
			s.Features[i] = feat
		}
		comm := cp.GetComment(f.Comment, true)
		if len(comm) > 0 {
			code += " //" + comm
		}
		s.FieldLines[i] = code
	}
	return size
}

// GetLineFeature 提取 struct field 的名称与类型作为特征
func GetLineFeature(code string) string {
	ps := strings.Fields(code)
	if len(ps) == 1 {
		return ps[0]
	}
	if strings.HasSuffix(ps[1], "json:\",inline\"") {
		return ps[0] + ":inline"
	}
	if strings.HasSuffix(ps[1], "xorm:\"extends\"") {
		return ps[0] + ":inline"
	}
	return ps[0] + ":" + ps[1]
}

// ReplaceModelFields
func ReplaceModelFields(cp *CodeParser, node *DeclNode, summary *ModelSummary) {
	var last ast.Node
	max := len(node.Fields) - 1
	first, lastField := ast.Node(node.Fields[0]), node.Fields[max]
	if lastField.Comment != nil {
		last = ast.Node(lastField.Comment)
	} else {
		last = ast.Node(lastField)
	}
	cp.AddReplace(first, last, summary.GetInnerCode())
}

// ReplaceSummary
func ReplaceSummary(summary, sub *ModelSummary) *ModelSummary {
	var features, lines []string
	find, sted := false, sub.GetSortedFeatures()
	for i, ft := range summary.Features {
		if !enums.InStringList(ft, sted) {
			features = append(features, ft)
			lines = append(lines, summary.FieldLines[i])
		} else if !find {
			subst := sub.GetSubstitute()
			features = append(features, subst)
			lines = append(lines, subst)
			find = true
			summary.IsChanged = true
		}
	}
	summary.Features, summary.FieldLines = features, lines
	return summary
}

// AddFormerMixins
func AddFormerMixins(fileName, nameSpace, alias string) []string {
	cp, err := NewFileParser(fileName)
	if err != nil {
		return nil
	}
	var mixinNames []string
	for _, node := range cp.AllDeclNode("type") {
		if len(node.Fields) == 0 {
			continue
		}
		name := node.GetName()
		if !strings.HasSuffix(name, "Mixin") {
			continue
		}
		summary := &ModelSummary{Import: nameSpace, Alias: alias}
		if alias == "" {
			alias = cp.GetPackage()
		}
		summary.Name = fmt.Sprintf("%s.%s", alias, name)
		_ = summary.ParseFields(cp, node)
		RegisterSubstitute(summary)
		mixinNames = append(mixinNames, summary.Name)
	}
	return mixinNames
}

// ScanAndUseMixins
func ScanAndUseMixins(summary, sub *ModelSummary, verbose bool) (needImport bool) {
	sted := sub.GetSortedFeatures()
	sorted := summary.GetSortedFeatures()
	// 函数 IsSubsetList(..., ..., true) 用于排除异名同构的Model
	if enums.IsSubsetList(sted, sorted, false) { // 正向替换
		summary = ReplaceSummary(summary, sub)
		if len(sorted) == len(sted) { //完全相等
			summary.IsExists = true
		}
		if sub.Import != "" {
			needImport = true
		}
		if verbose {
			fmt.Println(summary.Name, " <- ", sub.Name)
		}
	} else if strings.HasPrefix(sub.Name, "xquery.") {
		return // 早于反向替换，避免陷入死胡同
	} else if enums.IsSubsetList(sorted, sted, true) { // 反向替换
		ReplaceSummary(sub, summary)
		if verbose {
			fmt.Println(summary.Name, " -> ", sub.Name)
		}
	}
	return
}

// ParseAndMixinFile
func ParseAndMixinFile(fileName string, verbose bool) error {
	cp, err := NewFileParser(fileName)
	if err != nil {
		if verbose {
			fmt.Println(fileName, " error: ", err)
		}
		return err
	}
	var changed bool
	imports := make(map[string]string)
	for _, node := range cp.AllDeclNode("type") {
		if len(node.Fields) == 0 {
			continue
		}
		name := node.GetName()
		//if strings.Contains(cp.GetNodeCode(node), MODEL_EXTENDS) {
		//	continue // 避免重复处理 model
		//}

		summary := &ModelSummary{Name: name}
		_ = summary.ParseFields(cp, node)
		if summary.Isomorph() {
			summary.IsExists = true
		} else {
			for _, subName := range substituteNames {
				if subName == summary.Name {
					continue // 不要替换自己
				}
				if sub, ok := substituteModels[subName]; ok {
					if ScanAndUseMixins(summary, sub, verbose) {
						imports[sub.Import] = sub.Alias
					}
				}
			}
		}
		RegisterSubstitute(summary)
		if summary.IsChanged {
			changed = true
			ReplaceModelFields(cp, node, summary)
		}
	}
	if verbose {
		fmt.Println(fileName, " changed: ", changed, "\n")
	}
	if changed { // 加入相关的 mixin imports 并美化代码
		cs := cp.CodeSource
		if code, chg := cs.AltSource(); chg {
			cs.SetSource(code)
		}
		if cs, err = ResetImports(cs, imports); err != nil {
			return err
		}
		if err = cs.WriteTo(fileName); err != nil {
			return err
		}
	}
	return nil
}