package log

import (
	"os"

	slog "log"
)

const (
	prefix = "[autobuild] "
	flags  = slog.Ltime
)

var (
	normal   = slog.New(os.Stdout, prefix, flags)
	verbose1 = slog.New(os.Stdout, prefix, flags)
	verbose2 = slog.New(os.Stdout, prefix, flags)
)

// SetLevel sets the logging level. -1 is quiet, 0 is normal, anything higher is verbosity level.
func SetLevel(level int) {
	switch level {
	case -1:
		normal = nil
		fallthrough
	case 0:
		verbose1 = nil
		fallthrough
	case 1:
		verbose2 = nil
	}
}

func log(logger *slog.Logger, format string, v ...interface{}) {
	if logger == nil {
		return
	}
	if len(v) == 0 {
		logger.Println(format)
		return
	}
	logger.Printf(format+"\n", v...)
}

// Log is used for normal level logs.
func Log(format string, v ...interface{}) {
	log(normal, format, v...)
}

// V is used for verbose logs.
func V(format string, v ...interface{}) {
	log(verbose1, format, v...)
}

// VV is used for very verbose logs.
func VV(format string, v ...interface{}) {
	log(verbose2, format, v...)
}

// Fatalf is used to log and close the application.
func Fatalf(format string, v ...interface{}) {
	slog.Fatalf(format, v...)
}
