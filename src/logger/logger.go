package logger

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var Logger log.Logger

func InitLogger() {
	Logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	Logger = level.NewFilter(Logger, level.AllowAll())
	Logger = log.With(Logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
}
