package reverse

import (
	"path/filepath"

	"github.com/azhai/xgen/stuffs"
	"github.com/azhai/xgen/stuffs/web"
)

func SkelProject(outputDir, nameSpace string) (err error) {
	if err = stuffs.CopyFiles(outputDir, "./", "settings.hcl"); err != nil {
		return
	}
	data := map[string]any{"NameSpace": nameSpace, "ProjName": filepath.Base(nameSpace)}
	if err = web.GenFile(data, outputDir, "handlers", "record"); err != nil {
		return
	}
	if err = web.GenFile(data, outputDir, "cmd", "prepare"); err != nil {
		return
	}
	err = web.GenFile(data, outputDir, "cmd/serv", "main")
	return
}
