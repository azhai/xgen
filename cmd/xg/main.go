package main

import (
	"fmt"
	"sync"

	reverse "github.com/azhai/xgen"
	"github.com/azhai/xgen/cmd"
	"github.com/azhai/xgen/config"
	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/models"

	_ "github.com/arriqaaq/flashdb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/k0kubun/pp"
	_ "github.com/lib/pq"
	"github.com/manifoldco/promptui"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	opts, settings := cmd.GetOptions()
	models.Setup()
	var err error
	if opts.InterActive { // 采用交互模式，确定或修改部分配置
		if err = questions(settings); err != nil {
			fmt.Println("跳过，什么也没有做！")
			return // 到此结束
		}
	}

	// 生成models文件
	rver := reverse.NewGoReverser(settings.Reverse)
	if err = rver.GenModelInitFile("init"); err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	dbArgs := config.ReadArgs(true, nil)
	for _, cfg := range settings.Conns {
		if dbArgs.Size() > 0 && !dbArgs.Has(cfg.Key) {
			continue
		}
		wg.Add(1)
		go func(cfg dialect.ConnConfig) {
			defer wg.Done()
			currDir, isXorm := rver.SetOutDir(cfg.Key), true
			if opts.OnlyApplyMixins {
				isXorm = cfg.LoadDialect().IsXormDriver()
			} else {
				isXorm, err = rver.ExecuteReverse(cfg, opts.InterActive, opts.Verbose)
				if err != nil {
					panic(err)
				}
			}
			if isXorm {
				err = reverse.ApplyDirMixins(currDir, opts.Verbose)
				if err != nil {
					panic(err)
				}
			}
		}(cfg)
	}
	wg.Wait()
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
