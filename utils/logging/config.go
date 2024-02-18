package logging

import (
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevels 日志等级
var LogLevels = map[string]zapcore.Level{
	"debug":     zapcore.DebugLevel,
	"info":      zapcore.InfoLevel,
	"notice":    zapcore.InfoLevel, // zap中无notice等级
	"warn":      zapcore.WarnLevel,
	"warning":   zapcore.WarnLevel,  // warn的别名
	"err":       zapcore.ErrorLevel, // error的别名
	"error":     zapcore.ErrorLevel,
	"dpanic":    zapcore.DPanicLevel, // Develop环境会panic
	"panic":     zapcore.PanicLevel,  // 都会panic
	"fatal":     zapcore.FatalLevel,
	"critical":  zapcore.FatalLevel, // zap中无critical等级
	"emergency": zapcore.FatalLevel, // zap中无emergency等级
}

// GetZapLevel 转为zap的Level
func GetZapLevel(lvl string) (string, zapcore.Level) {
	if lvl = strings.ToLower(lvl); lvl != "" {
		if level, ok := LogLevels[lvl]; ok {
			return lvl, level
		}
	}
	return "", zapcore.DebugLevel
}

// GetLevelEnabler 级别过滤
func GetLevelEnabler(start, stop, min string) zapcore.LevelEnabler {
	var minLvl, startLvl, stopLvl zapcore.Level
	min, minLvl = GetZapLevel(min)
	start, startLvl = GetZapLevel(start)
	if start != min && startLvl.Enabled(minLvl) {
		startLvl = minLvl
	}
	if stop, stopLvl = GetZapLevel(stop); stop == "" {
		stopLvl = zapcore.FatalLevel
	}

	if stopLvl < startLvl {
		return nil
	} else if stopLvl == zapcore.FatalLevel {
		return zap.NewAtomicLevelAt(startLvl)
	} else {
		return zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= startLvl && lvl <= stopLvl
		})
	}
}

func NewEncoderConfig(timeFormat, levelFormat string) zapcore.EncoderConfig {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeCaller = nil
	if timeFormat == "" {
		ec.EncodeTime = nil
	} else {
		ec.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(timeFormat))
		}
	}
	switch strings.ToLower(levelFormat) {
	default:
		ec.EncodeLevel = nil
	case "cap", "capital":
		ec.EncodeLevel = zapcore.CapitalLevelEncoder
	case "color", "capcolor", "capitalcolor":
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case "low", "lower", "lowercase":
		ec.EncodeLevel = zapcore.LowercaseLevelEncoder
	case "lowcolor", "lowercolor", "lowercasecolor":
		ec.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	}
	return ec
}
