package xquery

import (
	"github.com/azhai/xgen/utils/logging"
	"go.uber.org/zap"
	"xorm.io/xorm/log"
)

type XormLogger struct {
	level   log.LogLevel
	showSQL bool
	*zap.SugaredLogger
}

func NewXormLogger(filename string) *XormLogger {
	xl := &XormLogger{level: log.LOG_INFO, showSQL: true}
	cfg := logging.SingleFileConfig("info", filename)
	xl.SugaredLogger = logging.NewLogger(cfg, "")
	return xl
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
