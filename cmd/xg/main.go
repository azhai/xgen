package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/azhai/gozzo/config"
	"github.com/azhai/gozzo/filesystem"
	reverse "github.com/azhai/xgen"
	"github.com/azhai/xgen/dialect"
	"github.com/azhai/xgen/rewrite"
	"github.com/k0kubun/pp"
	"github.com/manifoldco/promptui"
	_ "github.com/mattn/go-sqlite3"
)

var args struct {
	Pretty   *prettyCmd   `arg:"subcommand:pretty" help:"美化代码"`
	Mixin    *mixinCmd    `arg:"subcommand:mixin" help:"嵌入Mixins"`
	Skeleton *skeletonCmd `arg:"subcommand:skeleton" help:"生成新项目"`
	Config   string       `arg:"-c,--config" default:"settings.hcl" help:"配置文件路径"`
	Verbose  bool         `arg:"-v,--verbose" help:"输出详细信息"`
	reverseOpts
}

type prettyCmd struct {
	Dirs []string `arg:"positional"`
}

type mixinCmd struct {
}

type reverseOpts struct {
	IsInteract bool   `arg:"-i,--interact" default:false help:"交互模式"`
	OutputDir  string `arg:"-o,--out" default:"../example" help:"输出目录"`
	NameSpace  string `arg:"-n,--ns" default:"" help:"命名空间"`
}

type skeletonCmd struct {
	BinName string `arg:"-b,--bin" default:"serv" help:"二进制文件名"`
	IsForce bool   `arg:"-f,--force" default:false help:"覆盖文件"`
}

func init() {
	config.PrepareEnv(256)
	arg.MustParse(&args)
	config.ReadConfigFile(args.Config, args.Verbose, &args)
	if _, err := LoadConfigFile(false); err != nil {
		panic(err)
	}
}

func main() {
	if args.Pretty != nil { // 仅美化目录下的代码
		dirs := args.Pretty.Dirs
		if len(dirs) == 0 {
			prettifyDir(".")
			return
		}
		for _, dir := range dirs {
			prettifyDir(dir)
		}
		return
	}

	settings := GetDbSettings()
	if args.IsInteract { // 采用交互模式，确定或修改部分配置
		if err := questions(settings); err != nil {
			fmt.Println("跳过，什么也没有做！")
			return // 到此结束
		}
	}

	outputDir, _ := filepath.Abs(args.OutputDir)
	args.NameSpace = filepath.Base(args.OutputDir)
	settings.Reverse.OutputDir = filepath.Join(args.OutputDir, "models")
	settings.Reverse.NameSpace = fmt.Sprintf("%s/models", args.NameSpace)
	var skel *skeletonCmd
	if skel = args.Skeleton; skel != nil {
		_ = reverse.SkelProject(outputDir, args.NameSpace, skel.BinName, skel.IsForce)
	}
	rver := reverse.NewGoReverser(settings.Reverse)
	// 生成顶部目录下init单个文件
	if err := rver.GenModelInitFile("init"); err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	dbArgs := config.ReadArgs(true, nil)
	for _, cfg := range GetConnConfigs() {
		if dbArgs.Size() > 0 && !dbArgs.Has(cfg.Key) {
			continue
		}
		wg.Add(1)
		go func(rver *reverse.Reverser, cfg dialect.ConnConfig) {
			defer wg.Done()
			err := reverseDb(rver, cfg)
			if err != nil {
				fmt.Println("xx", err)
				panic(err)
			}
		}(rver.Clone(), cfg)
	}
	wg.Wait()

	fmt.Println("执行完成。")
	if skel != nil {
		_ = reverse.CheckProject(outputDir, args.NameSpace, skel.BinName)
	}
}

// prettifyDir 美化目录下的go代码文件
func prettifyDir(dir string) {
	files, err := filesystem.FindFiles(dir, ".go", "vendor/", ".git/")
	if err != nil {
		panic(err)
	}
	for filename := range files {
		fmt.Println("-", filename)
		_, _ = rewrite.RewriteGolangFile(filename, true)
	}
}

func reverseDb(rver *reverse.Reverser, cfg dialect.ConnConfig) (err error) {
	currDir, isXorm := rver.SetOutDir(cfg.Key), true
	if args.Mixin != nil { // 只进行Mixin嵌入
		isXorm = cfg.LoadDialect().IsXormDriver()
	} else { // 生成conn单个文件
		isXorm, err = rver.ExecuteReverse(cfg, args.IsInteract, args.Verbose)
		if err != nil {
			return
		}
	}
	if isXorm { // 生成models和queries多个文件
		err = reverse.ApplyDirMixins(currDir, args.Verbose)
	}
	return
}

// questions 交互式问题和接收回答
func questions(settings *DbSettings) (err error) {
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

	_, _ = pp.Println(settings.Reverse)
	prompt = promptui.Prompt{
		Label:     "使用以上配置，是否继续",
		IsConfirm: true,
	}
	_, err = prompt.Run()
	return
}
