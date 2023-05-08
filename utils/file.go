package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultDirMode  = 0o755
	DefaultFileMode = 0o644
)

// FileSize 检查文件是否存在及大小
// -1, false 不合法的路径
// 0, false 路径不存在
// -1, true 存在文件夹
// >=0, true 文件并存在
func FileSize(path string) (int64, bool) {
	if path == "" {
		return -1, false
	}
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return 0, false
	}
	var size = int64(-1)
	if info.IsDir() == false {
		size = info.Size()
	}
	return size, true
}

func CreateFile(path string) (fp *os.File, err error) {
	// create dirs if file not exists
	if dir := filepath.Dir(path); dir != "." {
		err = os.MkdirAll(dir, DefaultDirMode)
	}
	if err == nil {
		flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
		fp, err = os.OpenFile(path, flag, DefaultFileMode)
	}
	return
}

func OpenFile(path string, readonly, append bool) (fp *os.File, size int64, err error) {
	var exists bool
	size, exists = FileSize(path)
	if size < 0 {
		err = fmt.Errorf("Path is directory or illegal")
		return
	}
	if exists {
		flag := os.O_RDWR
		if readonly {
			flag = os.O_RDONLY
		} else if append {
			flag |= os.O_APPEND
		}
		fp, err = os.OpenFile(path, flag, DefaultFileMode)
	} else if readonly == false {
		fp, err = CreateFile(path)
	}
	return
}

// LineCount 使用 wc -l 计算有多少行
func LineCount(filename string) int {
	var num int
	filename, err := filepath.Abs(filename)
	if err != nil {
		return -1
	}
	out, err := exec.Command("wc", "-l", filename).Output()
	if err != nil {
		return -1
	}
	col := strings.SplitN(string(out), " ", 2)[0]
	if num, err = strconv.Atoi(col); err != nil {
		return -1
	}
	return num
}

// MkdirForFile 为文件路径创建目录
func MkdirForFile(path string) int64 {
	size, exists := FileSize(path)
	if size < 0 {
		return size
	}
	if !exists {
		dir := filepath.Dir(path)
		_ = os.MkdirAll(dir, DefaultDirMode)
	}
	return size
}

// FindFiles 遍历目录下的文件，递归方法
func FindFiles(dir, ext string, excls ...string) (map[string]os.FileInfo, error) {
	result := make(map[string]os.FileInfo)
	exclMatchers := NewGlobs(MapStrList(excls, func(s string) string {
		if strings.HasSuffix(s, string(filepath.Separator)) {
			return s + "*" // 匹配所有目录下所有文件和子目录
		}
		return s
	}, nil))
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil { // 终止
			return err
		} else if exclMatchers.MatchAny(path, false) { // 跳过
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if info.Mode().IsRegular() {
			if ext == "" || strings.HasSuffix(info.Name(), ext) {
				result[path] = info
			}
		}
		return nil
	})
	return result, err
}
