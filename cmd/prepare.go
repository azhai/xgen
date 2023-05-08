package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/azhai/xgen/config"
	"github.com/k0kubun/pp"
	"github.com/klauspost/cpuid/v2"
)

const (
	VERSION  = "1.6.0"
	MegaByte = 1024 * 1024
)

var (
	interActive      bool   // 交互模式
	onlyApplyMixins  bool   // 仅嵌入Mixins
	onlyPrettifyCode bool   // 仅美化代码
	onlyRunDemo      bool   // 运行样例代码
	onlyRunSkel      bool   // 生成新项目
	isForce          bool   // 覆盖文件
	skelOutDir       string // 新项目文件夹
)

// OptionConfig 自定义选项，一部分是命令行输入，一部分是配置文件解析
type OptionConfig struct {
	ExecAction  string
	InterActive bool
	NameSpace   string
	OutputDir   string
	IsForce     bool
	Verbose     bool
	Version     string `hcl:"version,optional" json:"version,omitempty"`
}

func init() {
	// 压舱石，阻止频繁GC
	ballast := make([]byte, 256*MegaByte)
	runtime.KeepAlive(ballast)

	if level := os.Getenv("GOAMD64"); level == "" {
		level = fmt.Sprintf("v%d", cpuid.CPU.X64Level())
		fmt.Printf("请设置环境变量 export GOAMD64=%s\n\n", level)
	}

	flag.BoolVar(&interActive, "i", false, "使用交互模式")
	flag.BoolVar(&onlyApplyMixins, "m", false, "仅嵌入Mixins")
	flag.StringVar(&skelOutDir, "o", "../example", "新项目文件夹")
	flag.BoolVar(&onlyPrettifyCode, "p", false, "仅美化代码")
	flag.BoolVar(&onlyRunDemo, "r", false, "运行样例代码")
	flag.BoolVar(&onlyRunSkel, "s", false, "生成新项目")
	flag.BoolVar(&isForce, "f", false, "覆盖文件")
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
	options.IsForce = isForce
	if onlyRunDemo {
		options.ExecAction = "demo"
	} else if onlyApplyMixins {
		options.ExecAction = "mixin"
	} else if onlyPrettifyCode {
		options.ExecAction = "pretty"
	} else if onlyRunSkel {
		options.ExecAction = "skel"
		options.NameSpace = filepath.Base(skelOutDir)
		options.OutputDir, _ = filepath.Abs(skelOutDir)
		settings.Reverse.OutputDir = filepath.Join(skelOutDir, "models")
		settings.Reverse.NameSpace = fmt.Sprintf("%s/models", options.NameSpace)
	}
	if options.Verbose = config.Verbose(); options.Verbose {
		fmt.Print("Options = ")
		pp.Println(options)
	}
	return options, settings
}
