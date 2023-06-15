package logging

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Output 输出配置
type Output struct {
	Start, Stop string
	OutPaths    []string
}

// LogConfig 日志配置
type LogConfig struct {
	zap.Config
	MinLevel   string
	LevelCase  string
	TimeFormat string
	Outputs    []Output
}

// NewLogger 根据配置产生记录器
func NewLogger(cfg *LogConfig, dir string) *zap.SugaredLogger {
	zl, err := cfg.BuildLogger(dir)
	if err == nil {
		return zl.Sugar()
	}
	panic(err)
}

// SingleFileConfig 使用单个文件的记录器
func SingleFileConfig(level, file string) *LogConfig {
	cfg := NewDefaultConfig()
	cfg.MinLevel = level
	if file != "" {
		cfg.Outputs = []Output{
			{Start: level, Stop: "fatal", OutPaths: []string{file}},
		}
	}
	return cfg
}

// NewDefaultConfig 默认配置，使用两个文件分别记录警告和错误
func NewDefaultConfig() *LogConfig {
	return &LogConfig{
		Config: zap.Config{
			Encoding:         "console",
			OutputPaths:      []string{},
			ErrorOutputPaths: []string{"stderr"}, // zap内部错误输出
		},
		MinLevel:   "debug",
		LevelCase:  "",
		TimeFormat: "2006-01-02 15:04:05",
		Outputs: []Output{
			{Start: "debug", Stop: "info", OutPaths: []string{"access.log"}},
			{Start: "warn", Stop: "fatal", OutPaths: []string{"error.log"}},
		},
	}
}

// BuildLogger 生成日志记录器
func (c *LogConfig) BuildLogger(dir string, opts ...zap.Option) (*zap.Logger, error) {
	if c.IsNop() {
		return zap.NewNop(), nil
	}
	if c.GetLevel().Enabled(zapcore.InfoLevel) {
		c.Development = true
		c.Sampling = nil
	}
	dir = strings.TrimSpace(dir)
	if cores := c.GetCores(dir); len(cores) > 1 {
		opts = append(opts, ReplaceCores(cores))
	}
	return c.Config.Build(opts...)
}

// IsNop 是否空日志
func (c *LogConfig) IsNop() bool {
	return len(c.Outputs) == 0 && len(c.OutputPaths) == 0
}

// GetLevel 当前日志的最低级别
func (c *LogConfig) GetLevel() zap.AtomicLevel {
	var level zapcore.Level
	c.MinLevel, level = GetZapLevel(c.MinLevel)
	c.Level = zap.NewAtomicLevelAt(level)
	return c.Level
}

// GetCores 产生记录器内核
func (c *LogConfig) GetCores(dir string) []zapcore.Core {
	var (
		cores []zapcore.Core
		ws    zapcore.WriteSyncer
		err   error
	)
	enc := c.GetEncoder()
	for _, out := range c.Outputs {
		enab := GetLevelEnabler(out.Start, out.Stop, c.MinLevel)
		if enab == nil || len(out.OutPaths) == 0 {
			continue
		}
		c.OutputPaths = GetLogPath(dir, out.OutPaths)
		if len(c.OutputPaths) == 0 || c.OutputPaths[0] == "/dev/null" {
			ws = zapcore.AddSync(io.Discard)
		} else if ws, _, err = zap.Open(c.OutputPaths...); err != nil {
			continue
		}
		cores = append(cores, zapcore.NewCore(enc, ws, enab))
	}
	return cores
}

// GetEncoder 根据编码配置设置日志格式
func (c *LogConfig) GetEncoder() zapcore.Encoder {
	c.Config.EncoderConfig = NewEncoderConfig(c.TimeFormat, c.LevelCase)
	if strings.ToLower(c.Encoding) == "json" {
		return zapcore.NewJSONEncoder(c.Config.EncoderConfig)
	}
	return zapcore.NewConsoleEncoder(c.Config.EncoderConfig)
}

// ReplaceCores 替换为多种输出的Core
func ReplaceCores(cores []zapcore.Core) zap.Option {
	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}

// GetLogPath 使用绝对路径
func GetLogPath(dir string, files []string) []string {
	if dir = strings.TrimSpace(dir); dir == "/dev/null" {
		return nil
	}
	for i, file := range files {
		files[i] = GetAbsPath(dir, file, true)
	}
	return files
}

// GetAbsPath 使用真实的绝对路径
func GetAbsPath(dir, file string, onlyFile bool) string {
	if dir == "" && strings.HasPrefix(file, "std") {
		return file
	}
	if strings.Contains(file, "%s") {
		file = fmt.Sprintf(dir, file)
	} else {
		file = fmt.Sprintf("%s/%s", dir, strings.TrimPrefix(file, "/"))
	}
	u, err := url.Parse(file)
	isFile := u.Scheme == "" || u.Scheme == "file"
	if err == nil && isFile == false && onlyFile {
		return file // 只能处理文件类型
	}
	path, _ := filepath.Abs(file)
	path = ignoreWinDisk(path)
	if isFile == false { // 重新拼接
		path = fmt.Sprintf("%s://%s?%s", u.Scheme, path, u.RawQuery)
	}
	return path
}
