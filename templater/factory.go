package templater

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/azhai/xgen/utils"
	"github.com/grsmv/inflect"
)

var (
	theFactory   = NewFactory("./", true).UpdateFuncs(defaultFuncs)
	defaultFuncs = template.FuncMap{
		"Lower":         strings.ToLower,
		"Upper":         strings.ToUpper,
		"Title":         strings.Title,
		"Camelize":      inflect.Camelize,
		"Underscore":    inflect.Underscore,
		"Singularize":   inflect.Singularize,
		"Pluralize":     inflect.Pluralize,
		"DiffPluralize": DiffPluralize,
	}
)

func init() {
	theFactory.Register("init", golangInitTemplate, nil)
	theFactory.Register("model", golangModelTemplate, nil)
	theFactory.Register("query", golangQueryTemplate, nil)
	theFactory.Register("xorm", golangXormTemplate, nil)
	theFactory.Register("redis", golangRedisTemplate, nil)
	theFactory.Register("flashdb", golangFlashdbTemplate, nil)
}

// DiffPluralize 如果复数形式和单数相同，人为增加后缀
func DiffPluralize(word, suffix string) string {
	words := inflect.Pluralize(word)
	if words == word {
		words += suffix
	}
	return words
}

// LoadTemplate 根据名称或路径加载模板对象
func LoadTemplate(nameOrPath string, funcs template.FuncMap) *template.Template {
	tmpl := theFactory.GetTemplate(nameOrPath, funcs)
	if tmpl == nil {
		tmpl = theFactory.RegisterFile(nameOrPath, funcs)
	}
	return tmpl
}

// RenderTemplate 使用数据渲染模板
func RenderTemplate(tmpl *template.Template, data any) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Factory 模板工厂，可设置模板目录用于扫描.tmpl文件
type Factory struct {
	inclSubFiles bool
	discoverDir  string
	presetTexts  map[string]string
	presetTmpls  map[string]*template.Template
	SharedFuncs  template.FuncMap
}

// NewFactory 创建工厂，可选的模板目录
func NewFactory(dir string, inclSubs bool) *Factory {
	if dir != "" {
		dir, _ = filepath.Abs(dir)
	}
	return &Factory{
		inclSubFiles: inclSubs,
		discoverDir:  dir,
		presetTexts:  make(map[string]string),
		presetTmpls:  make(map[string]*template.Template),
		SharedFuncs:  make(template.FuncMap),
	}
}

// UpdateFuncs 增加共享函数
func (f *Factory) UpdateFuncs(funcs template.FuncMap) *Factory {
	for name, fun := range funcs {
		f.SharedFuncs[name] = fun
	}
	return f
}

// AddSubs 加载子模板
func (f *Factory) AddSubs(tmpl *template.Template) *template.Template {
	if f.inclSubFiles && tmpl != nil {
		globPath := filepath.Join(f.discoverDir, "sub_*.tmpl")
		tmpl, _ = tmpl.ParseGlob(globPath)
	}
	return tmpl
}

// GetTemplate 根据名称获取模板对象，先找已注册模板，再找模板目录下文件
func (f *Factory) GetTemplate(name string, funcs template.FuncMap) *template.Template {
	if tmpl, ok := f.presetTmpls[name]; ok && len(funcs) == 0 {
		return f.AddSubs(tmpl)
	}
	if content, ok := f.presetTexts[name]; ok { // 残缺模板，funcs不齐全，只记录了内容
		tmpl := template.New(name).Funcs(f.SharedFuncs)
		if len(funcs) > 0 {
			tmpl = tmpl.Funcs(funcs)
		}
		tmpl = template.Must(tmpl.Parse(content))
		return f.AddSubs(tmpl)
	}
	if f.discoverDir == "" {
		return nil
	}
	filename := filepath.Join(f.discoverDir, name+".tmpl")
	if size, _ := utils.FileSize(filename); size > 0 {
		tmpl := f.RegisterFile(filename, funcs)
		return f.AddSubs(tmpl)
	}
	return nil
}

// Register 注册模板内容
func (f *Factory) Register(name, content string, funcs template.FuncMap) *template.Template {
	// Funcs adds the elements of the argument map to the template's function map.
	// It must be called before the template is parsed.
	tmpl := template.New(name).Funcs(f.SharedFuncs)
	if len(funcs) > 0 {
		tmpl = tmpl.Funcs(funcs)
	}
	f.presetTexts[name] = content
	var err error // 有可能所需的funcs不齐全，导致解析失败
	if tmpl, err = tmpl.Parse(content); err == nil {
		f.presetTmpls[name] = tmpl
	}
	return tmpl
}

// RegisterFile 注册模板文件
func (f *Factory) RegisterFile(filename string, funcs template.FuncMap) *template.Template {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	name := filepath.Base(filename)
	if extname := filepath.Ext(name); extname != "" {
		name = name[:len(name)-len(extname)]
	}
	return f.Register(name, string(fileData), funcs)
}

// Render 渲染指定模板
func (f *Factory) Render(name string, data any) ([]byte, error) {
	var tmpl *template.Template
	if tmpl = f.GetTemplate(name, nil); tmpl == nil {
		return nil, fmt.Errorf("cannot find the template named %s", name)
	}
	return RenderTemplate(tmpl, data)
}
