//go:build unix

package logging

// ignoreWinDisk 忽略Windows盘符
func ignoreWinDisk(absPath string) string {
	return absPath
}
