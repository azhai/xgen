package stuffs

import (
	"os"
	"path/filepath"

	"github.com/azhai/xgen/utils"
)

func CopyFiles(dest, src string, files map[string]string, force bool) (err error) {
	var content []byte
	for filename, toname := range files {
		if toname == "" {
			toname = filename
		}
		destFile := filepath.Join(dest, toname)
		size := utils.MkdirForFile(destFile)
		if !force && size > 0 { // 不要覆盖
			continue
		}
		srcFile := filepath.Join(src, filename)
		if content, err = os.ReadFile(srcFile); err != nil {
			return
		}
		err = os.WriteFile(destFile, content, utils.DefaultFileMode)
	}
	return
}
