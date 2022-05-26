package stuffs

import (
	"os"
	"path/filepath"

	"github.com/azhai/xgen/utils"
)

func CopyFiles(dest, src string, files ...string) (err error) {
	var content []byte
	for _, filename := range files {
		destFile := filepath.Join(dest, filename)
		size := utils.MkdirForFile(destFile)
		if size > 0 { //不要覆盖
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
