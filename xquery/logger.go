package xquery

import (
	"fmt"

	"github.com/azhai/xgen/utils/logging"
	"go.uber.org/zap"
	"xorm.io/xorm/log"
)

// XormLogger xorm日志
type XormLogger struct {
	level   log.LogLevel
	showSQL bool
	*zap.SugaredLogger
}

// NewXormLogger 创建日志
func NewXormLogger(filename string) *XormLogger {
	xl := &XormLogger{level: log.LOG_INFO, showSQL: true}
	cfg := logging.SingleFileConfig("info", filename)
	xl.SugaredLogger = logging.NewLogger(cfg, "")
	return xl
}

// AfterSQL implements ContextLogger
func (s *XormLogger) AfterSQL(ctx log.LogContext) {
	var sessionPart string
	v := ctx.Ctx.Value(log.SessionIDKey)
	if key, ok := v.(string); ok {
		sessionPart = fmt.Sprintf(" [%s]", key)
	}
	if ctx.ExecuteTime > 0 {
		s.Infof("[SQL]%s %s %v - %v", sessionPart, ctx.SQL, ctx.Args, ctx.ExecuteTime)
	} else {
		s.Infof("[SQL]%s %s %v", sessionPart, ctx.SQL, ctx.Args)
	}
}

// BeforeSQL implements ContextLogger
func (s *XormLogger) BeforeSQL(ctx log.LogContext) {
}

// Level implement log.Logger
func (s *XormLogger) Level() log.LogLevel {
	return s.level
}

// SetLevel implement log.Logger
func (s *XormLogger) SetLevel(l log.LogLevel) {
	s.level = l
	return
}

// ShowSQL implement log.Logger
func (s *XormLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		s.showSQL = true
		return
	}
	s.showSQL = show[0]
}

// IsShowSQL implement log.Logger
func (s *XormLogger) IsShowSQL() bool {
	return s.showSQL
}
