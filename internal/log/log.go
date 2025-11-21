package log

import (
	"fmt"
	stdlog "log"
	"os"
	"sync"
)

var (
	Quiet   bool
	Verbose bool
	Debug   bool
)

var configureStdLogOnce sync.Once

func configureStdLog() {
	stdlog.SetOutput(os.Stderr)
	stdlog.SetFlags(stdlog.LstdFlags | stdlog.Lshortfile)
}

func printToLogger(format string, args ...any) {
	configureStdLogOnce.Do(configureStdLog)
	if err := stdlog.Output(3, fmt.Sprintf(format, args...)); err != nil {
		panic(err)
	}
}

func printToStderr(format string, args ...any) {
	if _, err := fmt.Fprintf(os.Stderr, format+"\n", args...); err != nil {
		panic(err)
	}
}

func Debugf(format string, args ...any) {
	if Debug {
		printToLogger(format, args...)
	}
}

func Logf(format string, args ...any) {
	if Debug {
		printToLogger(format, args...)
		return
	}
	if Quiet {
		return
	}
	if !Verbose {
		return
	}
	printToStderr(format, args...)
}

func Infof(format string, args ...any) {
	if Debug {
		printToLogger(format, args...)
		return
	}
	if Quiet {
		return
	}
	printToStderr(format, args...)
}

func Warnf(format string, args ...any) {
	if Debug {
		printToLogger(format, args...)
		return
	}
	if Quiet {
		return
	}
	printToStderr("WARNING: "+format, args...)
}

func Errorf(format string, args ...any) {
	if Debug {
		printToLogger(format, args...)
		return
	}
	printToStderr("ERROR: "+format, args...)
}
