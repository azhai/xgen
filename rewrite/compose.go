package rewrite

import (
	"bytes"
	"fmt"
	"go/ast"
	"sort"
	"strings"

	"github.com/azhai/gozzo/filesystem"
	"github.com/azhai/gozzo/match"
	"github.com/azhai/xgen/utils"
	"github.com/pkg/errors"
)

const (
	AdaptivePkgName = "#"
	ModelExtends    = "`json:\",inline\" xorm:\"extends\"`"
)

var (
	MixinWildcards  = []string{"*Core", "*Mixin"} // 可注册的Mixin类名形式
	globaltComposer = GlobalComposer()            // 公共Mixins嵌入器
)

// Composer 是一个Model分析和嵌入工具
type Composer struct {
	subNames  []string
	subModels map[string]*ModelSummary
	Global    *Composer
}

// NewComposer 创造嵌入工具，带一个全局公共嵌入器
func NewComposer() *Composer {
	return &Composer{
		subModels: make(map[string]*ModelSummary),
		Global:    globaltComposer,
	}
}

// GlobalComposer 公共嵌入器，带有xquery的两个mixins
func GlobalComposer() *Composer {
	cps := &Composer{
		subNames:  []string{"xq.NestedMixin", "xq.TimeMixin"},
		subModels: make(map[string]*ModelSummary),
	}
	cps.subModels["xq.NestedMixin"] = &ModelSummary{
		Name:   "xq.NestedMixin",
		Import: "github.com/azhai/xgen/xquery",
		Alias:  "xq",
		FieldLines: []string{
			"Lft   int `json:\"lft\" xorm:\"notnull default 0 comment('左边界') INT(10)\"`           // 左边界",
			"Rgt   int `json:\"rgt\" xorm:\"notnull default 0 comment('右边界') index INT(10)\"`     // 右边界",
			"Depth int `json:\"depth\" xorm:\"notnull default 1 comment('高度') index TINYINT(3)\"` // 高度",
		},
	}
	cps.subModels["xq.TimeMixin"] = &ModelSummary{
		Name:   "xq.TimeMixin",
		Import: "github.com/azhai/xgen/xquery",
		Alias:  "xq",
		FieldLines: []string{
			"CreatedAt time.Time `json:\"created_at\" xorm:\"created comment('创建时间') TIMESTAMP\"`       // 创建时间",
			"UpdatedAt time.Time `json:\"updated_at\" xorm:\"updated comment('更新时间') TIMESTAMP\"`       // 更新时间",
			"DeletedAt time.Time `json:\"deleted_at\" xorm:\"deleted comment('删除时间') index TIMESTAMP\"` // 删除时间",
		},
	}
	return cps
}

// RegisterGlobalSubstitute 注册可替换Model
func (c *Composer) RegisterGlobalSubstitute(sub *ModelSummary) {
	c.Global.RegisterSubstitute(sub)
}

// RegisterSubstitute 注册可替换Model
func (c *Composer) RegisterSubstitute(sub *ModelSummary) {
	if sub != nil && !sub.IsExists {
		if _, ok := c.subModels[sub.Name]; !ok {
			c.subNames = append(c.subNames, sub.Name)
		}
		c.subModels[sub.Name] = sub
	}
}

// RemoveSubstitute 删除可替换Model
func (c *Composer) RemoveSubstitute(name string) {
	if _, ok := c.subModels[name]; ok {
		c.subModels[name] = nil
	}
}

// SubstituteSummary 替换和改写Model
func (c *Composer) SubstituteSummary(summary *ModelSummary, verbose bool) []*ModelSummary {
	var subs []*ModelSummary
	if c.Global != nil && len(c.Global.subNames) > 0 {
		subs = c.Global.SubstituteSummary(summary, verbose)
	}
	size, checks := len(c.subNames), make(map[string]string)
	for i := size - 1; i >= 0; i-- { // 倒序，优先使用更大的subModel
		subName := c.subNames[i]
		if subName == summary.Name {
			continue // 不要替换自己
		}
		if _, ok := checks[subName]; ok {
			continue // 不要替换子集的子集
		}
		if sub, ok := c.subModels[subName]; ok {
			match, needImport := summary.ScanAndUseMixins(sub, false, verbose)
			if match {
				summary.Children = append(summary.Children, subName)
				for _, child := range sub.Children {
					checks[child] = sub.Name
				}
			}
			if needImport {
				subs = append(subs, sub)
			}
		}
	}
	return subs
}

// ParseAndMixinFile 使用Mixin改写文件
func (c *Composer) ParseAndMixinFile(filename string, verbose bool) error {
	cp, err := NewFileParser(filename)
	if err != nil {
		if verbose {
			fmt.Println(filename, " error: ", err)
		}
		return err
	}
	var changed bool
	imports := make(map[string]string)
	for _, node := range cp.AllDeclNode("type") {
		if len(node.Fields) == 0 {
			continue
		}
		summary := &ModelSummary{Name: node.GetName()}
		_ = summary.ParseFields(cp, node)
		if summary.Isomorphic() {
			summary.IsExists = true
		} else {
			for _, sub := range c.SubstituteSummary(summary, verbose) {
				if sub.Alias == "#" { // 重复的mixin会产生#别名
					sub.Alias = ""
				}
				imports[sub.Import] = sub.Alias
			}
		}
		c.RegisterSubstitute(summary)
		if summary.IsChanged {
			changed = true
			ReplaceModelFields(cp, node, summary)
		}
	}
	if changed { // 加入相关的 mixin imports 并美化代码
		err = cp.ResetImports(filename, imports)
	}
	if verbose {
		if changed {
			fmt.Println("+", filename)
		} else {
			fmt.Println("-", filename)
		}
	}
	if err != nil {
		err = errors.Wrap(err, "ResetImports failed "+filename)
	}
	return err
}

// GetLineFeature 提取 struct field 的名称与类型作为特征
func GetLineFeature(code string) string {
	ps := strings.Fields(code)
	if len(ps) == 1 {
		return ps[0]
	}
	if strings.Contains(ps[1], "json:\",inline\"") {
		return ps[0] + ":inline"
	}
	if strings.Contains(ps[1], "xorm:\"extends\"") {
		return ps[0] + ":inline"
	}
	return ps[0] + ":" + ps[1]
}

// ModelSummary Model摘要
type ModelSummary struct {
	Name           string
	Substitute     string
	Import, Alias  string
	Features       []string
	sortedFeatures []string
	FieldLines     []string
	Children       []string
	IsChanged      bool
	IsExists       bool // 同构Model已存在
}

// GetInnerCode 找出 model 内部代码，即在 {} 里面的内容
func (s *ModelSummary) GetInnerCode() string {
	var buf bytes.Buffer
	for _, line := range s.FieldLines {
		buf.WriteString(fmt.Sprintf("\t%s\n", line))
	}
	return buf.String()
}

// GetSortedFeatures 找出 model 的所有特征并排序
func (s *ModelSummary) GetSortedFeatures() []string {
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

// Isomorphic 已经是其他Model的同构体，没有嵌入的空间
func (s *ModelSummary) Isomorphic() bool {
	features := s.GetSortedFeatures()
	return len(features) == 1 && strings.HasSuffix(features[0], ":inline")
}

// GetSubstitute 使用inline tag代替原来的那些字段，prefix可以是*星号
func (s *ModelSummary) GetSubstitute(prefix string) string {
	if s.Substitute == "" {
		s.Substitute = fmt.Sprintf("%s%s %s", prefix, s.Name, ModelExtends)
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
			code += " //" + match.TruncateText(comm, 50)
		}
		s.FieldLines[i] = code
	}
	return size
}

// ReplaceSummary 使用subModel重写当前Model
func (s *ModelSummary) ReplaceSummary(sub *ModelSummary) bool {
	var features, lines []string
	find, sted := false, sub.GetSortedFeatures()
	for i, ft := range s.Features {
		if !match.InStringList(ft, sted) {
			features = append(features, ft)
			lines = append(lines, s.FieldLines[i])
		} else if !find {
			subst := sub.GetSubstitute("")
			features = append(features, GetLineFeature(subst))
			lines = append(lines, subst)
			find = true
			s.IsChanged = true
		}
	}
	s.Features, s.FieldLines = features, lines
	return s.IsChanged
}

// ScanAndUseMixins 扫描和使用Mixin
func (s *ModelSummary) ScanAndUseMixins(sub *ModelSummary, deep, verbose bool) (matched, needImport bool) {
	sted := sub.GetSortedFeatures()
	sorted := s.GetSortedFeatures()
	// 函数 IsSubsetList(..., ..., true) 用于排除异名同构的Model
	if match.IsSubsetList(sted, sorted, false) { // 正向替换
		matched = s.ReplaceSummary(sub)
		if len(sorted) == len(sted) { // 完全相等
			s.IsExists = true
		}
		if sub.Import != "" {
			needImport = true
		}
		if verbose {
			fmt.Println("*", s.Name, " <- ", sub.Name)
		}
	}
	if deep == false {
		return
	}
	if strings.HasPrefix(sub.Name, "xq.") || strings.HasPrefix(sub.Name, "xquery.") {
		return // 早于反向替换，避免陷入死胡同
	} else if match.IsSubsetList(sorted, sted, true) { // 反向替换
		sub.ReplaceSummary(s)
		if verbose {
			fmt.Println("*", s.Name, " -> ", sub.Name)
		}
	}
	return
}

// ReplaceModelFields 将Mixin写入到Model内，替代它的部分字段
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

// AddFormerMixins 将文件中符合的Mixin类先行注册
func AddFormerMixins(cps *Composer, filename, nameSpace, alias string) []string {
	cp, err := NewFileParser(filename)
	if err != nil {
		return nil
	}
	var mixinNames []string
	// 以Core或Mixin结尾的类才会嵌入Model中
	for _, node := range cp.FindDeclNode("type", MixinWildcards...) {
		if len(node.Fields) == 0 {
			continue
		}
		summary := &ModelSummary{Import: nameSpace, Alias: alias}
		if alias == AdaptivePkgName {
			alias = cp.GetPackage()
		}
		name := node.GetName()
		if alias == "" {
			summary.Name = name
		} else {
			summary.Name = fmt.Sprintf("%s.%s", alias, name)
		}
		_ = summary.ParseFields(cp, node)
		cps.RegisterSubstitute(summary)
		mixinNames = append(mixinNames, summary.Name)
	}
	return mixinNames
}

// PrepareMixins 扫描目录中的可注册Mixin，并声明为某个NameSpace下
func PrepareMixins(mixinDir, mixinNS string) (mixinNames []string) {
	if _, isExists := utils.FileSize(mixinDir); !isExists {
		return
	}
	files, _ := filesystem.FindFiles(mixinDir, ".go")
	for filename := range files {
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		newNames := AddFormerMixins(globaltComposer, filename, mixinNS, AdaptivePkgName)
		mixinNames = append(mixinNames, newNames...)
	}
	return
}
