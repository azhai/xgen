package utils

import (
	"os"
	"path/filepath"
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
	size := int64(-1)
	if info.IsDir() == false {
		size = info.Size()
	}
	return size, true
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
