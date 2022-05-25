package rewrite

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"strings"

	"github.com/azhai/xgen/utils"
	"golang.org/x/tools/imports"
)

// TrimComment 去掉注释两边的空白
func TrimComment(c string) string {
	c = strings.TrimSpace(c)
	if strings.HasPrefix(c, "//") {
		c = strings.TrimSpace(c[2:])
	}
	return c
}

// FormatGolangCode 格式化代码，如果出错返回原内容
func FormatGolangCode(src []byte) ([]byte, error) {
	_src, err := format.Source(src)
	if err == nil {
		src = _src
	}
	return src, err
}

// SaveCodeToFile 将go代码保存到文件
func SaveCodeToFile(filename string, codeText []byte) ([]byte, error) {
	utils.MkdirForFile(filename)
	err := ioutil.WriteFile(filename, codeText, utils.DefaultFileMode)
	return codeText, err
}

// RewriteGolangFile 读出来go代码，重新写入文件
func RewriteGolangFile(filename string, cleanImports bool) (changed bool, err error) {
	var srcCode, dstCode []byte
	if srcCode, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	dstCode, err = writeGolangFile(filename, srcCode, cleanImports)
	if bytes.Compare(srcCode, dstCode) != 0 {
		changed = true
	}
	return
}

func writeGolangFile(filename string, codeText []byte, cleanImports bool) ([]byte, error) {
	// Formart/Prettify the code 格式化代码
	srcCode, err := FormatGolangCode(codeText)
	if err != nil {
		//fmt.Println(err)
		//fmt.Println(string(codeText))
		return codeText, err
	}
	if cleanImports { // 清理 import
		cs := NewCodeSource()
		cs.SetSource(srcCode)
		if cs.CleanImports() > 0 {
			if srcCode, err = cs.GetContent(); err != nil {
				return srcCode, err
			}
		}
	}
	if _, err = SaveCodeToFile(filename, srcCode); err != nil {
		return srcCode, err
	}
	// Split the imports in two groups: go standard and the third parts 分组排序引用包
	var dstCode []byte
	dstCode, err = imports.Process(filename, srcCode, nil)
	if err != nil {
		return srcCode, err
	}
	return SaveCodeToFile(filename, dstCode)
}

// WriteGolangFilePrettify 美化并输出go代码到文件
func WriteGolangFilePrettify(filename string, codeText []byte) ([]byte, error) {
	return writeGolangFile(filename, codeText, false)
}

// WriteGolangFilePrettify 美化和整理导入，并输出go代码到文件
func WriteGolangFileCleanImports(filename string, codeText []byte) ([]byte, error) {
	return writeGolangFile(filename, codeText, true)
}

// RewritePackage 将包中的Go文件格式化，如果提供了pkgname则用作新包名
func RewritePackage(pkgpath, pkgname string) error {
	if pkgname != "" {
		// TODO: 替换包名
	}
	files, err := utils.FindFiles(pkgpath, ".go")
	if err != nil {
		return err
	}
	var content []byte
	for filename := range files {
		content, err = ioutil.ReadFile(filename)
		if err != nil {
			break
		}
		_, err = WriteGolangFilePrettify(filename, content)
		if err != nil {
			break
		}
	}
	return err
}

// RewriteWithImports 注入导入声明
func RewriteWithImports(pkg string, source []byte, imports map[string]string) (*CodeSource, error) {
	cs := NewCodeSource()
	if err := cs.SetPackage(pkg); err != nil {
		return cs, err
	}
	// 添加可能引用的包，后面再尝试删除不一定会用的包
	for imp, alias := range imports {
		cs.AddImport(imp, alias)
	}
	if err := cs.AddCode(source); err != nil {
		return cs, err
	}
	for imp, alias := range imports {
		cs.DelImport(imp, alias)
	}
	return cs, nil
}
