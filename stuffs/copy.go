package stuffs

import (
	"os"
	"path/filepath"

	"github.com/azhai/xgen/utils"
)

func CopyFiles(dest, src string, files ...string) (err error) {
	var content []byte
	for _, filename := range files {
		srcFile := filepath.Join(src, filename)
		if content, err = os.ReadFile(srcFile); err != nil {
			return
		}
		destFile := filepath.Join(dest, filename)
		utils.MkdirForFile(destFile)
		err = os.WriteFile(destFile, content, utils.DefaultFileMode)
	}
	return
}
