package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/azhai/gozzo/config"
	"github.com/azhai/xgen/cmd"
	"github.com/k0kubun/pp"
)

var (
	skelOutDir string // 新项目文件夹
	options    = new(CommandOptions)
)

// CommandOptions 命令行自定义选项
type CommandOptions struct {
	ExecAction       string
	OnlyApplyMixins  bool // 仅嵌入Mixins
	OnlyPrettifyCode bool // 仅美化代码
	OnlyRunDemo      bool // 运行样例代码
	OnlyRunSkel      bool // 生成新项目
	InterActive      bool // 交互模式
	IsForce          bool // 覆盖文件
	NameSpace        string
	OutputDir        string
}

func init() {
	config.PrepareEnv(256)
	flag.StringVar(&skelOutDir, "o", "../example", "新项目文件夹")
	flag.BoolVar(&options.InterActive, "i", false, "使用交互模式")
	flag.BoolVar(&options.IsForce, "f", false, "覆盖文件")
	flag.BoolVar(&options.OnlyApplyMixins, "m", false, "仅嵌入Mixins")
	flag.BoolVar(&options.OnlyPrettifyCode, "p", false, "仅美化代码")
	flag.BoolVar(&options.OnlyRunDemo, "r", false, "运行样例代码")
	flag.BoolVar(&options.OnlyRunSkel, "s", false, "生成新项目")
	config.PrepareFlags()

	if _, err := cmd.LoadConfigFile(false); err != nil {
		panic(err)
	}
}

// ParseOptions 解析配置文件和命令行命名参数
func ParseOptions() (*CommandOptions, *cmd.DbSettings) {
	settings := cmd.GetDbSettings()

	if options.OnlyRunDemo {
		options.ExecAction = "demo"
	} else if options.OnlyApplyMixins {
		options.ExecAction = "mixin"
	} else if options.OnlyPrettifyCode {
		options.ExecAction = "pretty"
	} else if options.OnlyRunSkel {
		options.ExecAction = "skel"
		options.NameSpace = filepath.Base(skelOutDir)
		options.OutputDir, _ = filepath.Abs(skelOutDir)
		settings.Reverse.OutputDir = filepath.Join(skelOutDir, "models")
		settings.Reverse.NameSpace = fmt.Sprintf("%s/models", options.NameSpace)
	}

	if config.Verbose() {
		fmt.Print("Options = ")
		_, _ = pp.Println(options)
	}
	return options, settings
}
