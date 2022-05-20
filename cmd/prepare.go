package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/azhai/xgen/config"
	"github.com/k0kubun/pp"
	"github.com/klauspost/cpuid/v2"
)

const (
	VERSION = "1.5.0"
)

var (
	execAction  string // 执行操作
	interActive bool   // 交互模式
)

// OptionConfig 自定义选项，一部分是命令行输入，一部分是配置文件解析
type OptionConfig struct {
	ExecAction  string
	InterActive bool
	Verbose     bool
	Version     string `hcl:"version,optional" json:"version,omitempty"`
}

func init() {
	if level := os.Getenv("GOAMD64"); level == "" {
		level = fmt.Sprintf("v%d", cpuid.CPU.X64Level())
		fmt.Printf("请设置环境变量 export GOAMD64=%s\n\n", level)
	}

	flag.StringVar(&execAction, "e", "gen", "执行操作")
	flag.BoolVar(&interActive, "i", false, "使用交互模式")
	config.Setup()
}

// GetOptions 解析配置文件和命令行命名参数
func GetOptions() (*OptionConfig, *config.RootConfig) {
	options := new(OptionConfig)
	settings, err := config.ReadConfigFile(options)
	if err != nil {
		panic(err)
	}
	if options.Version == "" {
		options.Version = VERSION
	}
	options.InterActive = interActive
	options.ExecAction = strings.ToLower(execAction)
	if options.Verbose = config.Verbose(); options.Verbose {
		fmt.Print("Options = ")
		pp.Println(options)
	}
	return options, settings
}
