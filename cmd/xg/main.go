package main

import (
	"flag"
	"fmt"
	"sync"

	reverse "github.com/azhai/xgen"
	"github.com/azhai/xgen/cmd"
	"github.com/azhai/xgen/config"
	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/rewrite"
	"github.com/azhai/xgen/utils"

	_ "github.com/arriqaaq/flashdb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/k0kubun/pp"
	_ "github.com/lib/pq"
	"github.com/manifoldco/promptui"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	opts, settings := cmd.GetOptions()
	if opts.ExecAction != "" {
		fmt.Println(opts.ExecAction, "...")
	}

	if opts.ExecAction == "demo" {
		runTheDemo()
		return
	} else if opts.ExecAction == "pretty" { // 仅美化目录下的代码
		if flag.NArg() == 0 {
			prettifyDir(".")
			return
		}
		for _, dir := range flag.Args() {
			prettifyDir(dir)
		}
		return
	}

	var err error
	if opts.InterActive { // 采用交互模式，确定或修改部分配置
		if err = questions(settings); err != nil {
			fmt.Println("跳过，什么也没有做！")
			return // 到此结束
		}
	}

	if opts.ExecAction == "skel" {
		reverse.SkelProject(opts.OutputDir, opts.NameSpace, "serv")
	}
	rver := reverse.NewGoReverser(settings.Reverse)
	// 生成顶部目录下init单个文件
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
		go func(rver *reverse.Reverser, cfg dialect.ConnConfig) {
			defer wg.Done()
			err := reverseDb(rver, cfg, opts)
			if err != nil {
				panic(err)
			}
		}(rver.Clone(), cfg)
	}
	wg.Wait()

	fmt.Println("执行完成。")
	reverse.CheckProject(opts.OutputDir, opts.NameSpace)
}

// prettifyDir 美化目录下的go代码文件
func prettifyDir(dir string) {
	files, err := utils.FindFiles(dir, ".go", "vendor/", ".git/")
	if err != nil {
		panic(err)
	}
	for filename := range files {
		fmt.Println("-", filename)
		rewrite.RewriteGolangFile(filename, true)
	}
}

func reverseDb(rver *reverse.Reverser, cfg dialect.ConnConfig, opts *cmd.OptionConfig) (err error) {
	currDir, isXorm := rver.SetOutDir(cfg.Key), true
	if opts.ExecAction == "mixin" { // 只进行Mixin嵌入
		isXorm = cfg.LoadDialect().IsXormDriver()
	} else { // 生成conn单个文件
		isXorm, err = rver.ExecuteReverse(cfg, opts.InterActive, opts.Verbose)
		if err != nil {
			return
		}
	}
	if isXorm { // 生成models和queries多个文件
		err = reverse.ApplyDirMixins(currDir, opts.Verbose)
	}
	return
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
