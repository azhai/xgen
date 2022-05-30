package reverse

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/azhai/xgen/stuffs"
	"github.com/azhai/xgen/stuffs/web"
	"github.com/azhai/xgen/utils"
)

// CheckProject 检查go mod相关文件，并给出编译提示
func CheckProject(outputDir, nameSpace, binName string) (err error) {
	data := map[string]any{"NameSpace": nameSpace, "ProjName": filepath.Base(nameSpace)}
	if err = web.GenFile(data, filepath.Join(outputDir, "cmd", binName), "main"); err != nil {
		return
	}
	modfile := filepath.Join(outputDir, "go.mod")
	if _, ok := utils.FileSize(modfile); ok {
		return
	}
	fmt.Print("\n请接着执行下面操作，初始化go项目：\n")
	fmt.Print("export GOPROXY=https://goproxy.cn\n")
	fmt.Printf("cd %s\n", outputDir)
	fmt.Printf("go mod init \"%s\"\n", nameSpace)
	fmt.Print("go mod tidy\nmake\n")
	return
}

// SkelProject 生成一个项目的骨架
func SkelProject(outputDir, nameSpace, binName string, force bool) (err error) {
	files := map[string]string{"settings.hcl.example": "settings.hcl", ".gitignore": "", "Makefile": ""}
	if err = stuffs.CopyFiles(outputDir, "./", files, force); err != nil {
		return
	}
	filename := filepath.Join(outputDir, "Makefile")
	sedcmd := fmt.Sprintf("/^COMMANDS = /cCOMMANDS = %s", binName) // 不需要加单引号
	if err = exec.Command("/usr/bin/sed", "-i", sedcmd, filename).Run(); err != nil {
		return
	}
	data := map[string]any{"NameSpace": nameSpace, "ProjName": filepath.Base(nameSpace)}
	if err = web.GenFile(data, filepath.Join(outputDir, "handlers"), "record"); err != nil {
		return
	}
	err = web.GenFile(data, filepath.Join(outputDir, "cmd"), "prepare")
	return
}
