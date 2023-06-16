package xgen_test

import (
	"testing"
	"time"

	"github.com/azhai/xgen/utils/logging"
)

var (
	cfg    = logging.SingleFileConfig("info", "./logs/access.log")
	logger = logging.NewLogger(cfg, "")
)

func NowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Test11Info(t *testing.T) {
	logger.Info("999 888")
	logger.Errorf("now is %s", NowTime())
	// assert.NoError(t, err)
}
