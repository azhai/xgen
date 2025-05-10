package web

import (
	"fmt"
	"path/filepath"

	"github.com/azhai/xgen/rewrite"
	"github.com/azhai/xgen/templater"
)

var tmpls = templater.NewFactory("./stuffs/web/", false)

func GenFile(data map[string]any, dest string, names ...string) (err error) {
	var content []byte
	for _, name := range names {
		if content, err = tmpls.Render(name, data); err != nil {
			return
		}
		filename := filepath.Join(dest, name+".go")
		fmt.Println(">", filename)
		if _, err = rewrite.WriteGolangFilePrettify(filename, content); err != nil {
			return
		}
	}
	return
}
