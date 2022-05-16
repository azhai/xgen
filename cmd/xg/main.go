package main

import (
	"fmt"
	"os"

	reverse "github.com/azhai/xgen"
	"github.com/azhai/xgen/cmd"
	"github.com/azhai/xgen/config"
	"github.com/azhai/xgen/models"

	_ "github.com/arriqaaq/flashdb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/k0kubun/pp"
	_ "github.com/lib/pq"
	"github.com/manifoldco/promptui"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	options, settings := cmd.GetOptions()
	models.Setup()
	var err error
	if options.InterActive {
		if err = questions(settings); err != nil {
			fmt.Println("跳过，什么也没有做！")
			os.Exit(0)
		}
	}
	rver := reverse.NewGoReverser(settings.Reverse)
	if err = rver.GenModelInitFile("init"); err != nil {
		panic(err)
	}
	for _, cfg := range settings.Conns {
		err = rver.ExecuteReverse(cfg, options.InterActive, options.Verbose)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("执行完成。")
}

// questions 交互式问题和接收回答
func questions(settings *config.RootConfig) (err error) {
	prompt := promptui.Prompt{
		Label:     "使用交互模式生成多组Model文件，开始",
		IsConfirm: true,
	}
	if _, err = prompt.Run(); err != nil { // 没有选择Yes
		return
	}

	val := settings.Reverse.OutputDir
	prompt = promptui.Prompt{
		Label:   "Model文件输出目录",
		Default: val,
	}
	if val, err = prompt.Run(); err != nil {
		return
	}
	settings.Reverse.OutputDir = val

	val = settings.Reverse.NameSpace
	prompt = promptui.Prompt{
		Label:   "Model的完整包名路径",
		Default: val,
	}
	if val, err = prompt.Run(); err != nil {
		return
	}
	settings.Reverse.NameSpace = val

	pp.Println(settings.Reverse)
	prompt = promptui.Prompt{
		Label:     "使用以上配置，是否继续",
		IsConfirm: true,
	}
	_, err = prompt.Run()
	return
}
