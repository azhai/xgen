package rewrite

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"
)

const (
	DEFAULT_DIR_MODE  = 0o755
	DEFAULT_FILE_MODE = 0o644
)

// FindFiles 遍历目录下的文件，递归
func FindFiles(dir, ext string) (map[string]os.FileInfo, error) {
	result := make(map[string]os.FileInfo)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Mode().IsRegular() {
			if ext == "" || strings.HasSuffix(info.Name(), ext) {
				result[path] = info
			}
		}
		return err
	})
	return result, err
}

// FormatGolangCode 格式化代码，如果出错返回原内容
func FormatGolangCode(src []byte) ([]byte, error) {
	_src, err := format.Source(src)
	if err == nil {
		src = _src
	}
	return src, err
}

func WriteCodeFile(fileName string, codeText []byte) ([]byte, error) {
	err := ioutil.WriteFile(fileName, codeText, DEFAULT_FILE_MODE)
	return codeText, err
}

func writeGolangFile(fileName string, codeText []byte, cleanImports bool) ([]byte, error) {
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
	if _, err = WriteCodeFile(fileName, srcCode); err != nil {
		return srcCode, err
	}
	// Split the imports in two groups: go standard and the third parts 分组排序引用包
	var dstCode []byte
	dstCode, err = imports.Process(fileName, srcCode, nil)
	if err != nil {
		return srcCode, err
	}
	return WriteCodeFile(fileName, dstCode)
}

// WriteGolangFilePrettify 美化并输出go代码到文件
func WriteGolangFilePrettify(fileName string, codeText []byte) ([]byte, error) {
	return writeGolangFile(fileName, codeText, false)
}

// WriteGolangFilePrettify 美化和整理导入，并输出go代码到文件
func WriteGolangFileCleanImports(fileName string, codeText []byte) ([]byte, error) {
	return writeGolangFile(fileName, codeText, true)
}

// PrettifyGolangFile 格式化Go文件
func PrettifyGolangFile(fileName string, cleanImports bool) (changed bool, err error) {
	var srcCode, dstCode []byte
	if srcCode, err = ioutil.ReadFile(fileName); err != nil {
		return
	}
	dstCode, err = writeGolangFile(fileName, srcCode, cleanImports)
	if bytes.Compare(srcCode, dstCode) != 0 {
		changed = true
	}
	return
}

// RewritePackage 将包中的Go文件格式化，如果提供了pkgname则用作新包名
func RewritePackage(pkgpath, pkgname string) error {
	if pkgname != "" {
		// TODO: 替换包名
	}
	files, err := FindFiles(pkgpath, ".go")
	if err != nil {
		return err
	}
	var content []byte
	for fileName := range files {
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			break
		}
		_, err = WriteGolangFilePrettify(fileName, content)
		if err != nil {
			break
		}
	}
	return err
}
