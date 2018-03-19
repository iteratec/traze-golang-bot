package logging

import (
	"github.com/op/go-logging"
)

var Log = logging.MustGetLogger("example")
var format = logging.MustStringFormatter(
	`%{color} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// initializes logging framework
func InitLogging() {
	logging.SetFormatter(format)
}
