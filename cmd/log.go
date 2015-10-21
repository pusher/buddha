package main

import (
	stdlog "log"
	"os"
)

// global logger
var log = Log{}

// leveled logger
type Log struct{}

func (l Log) Printf(format string, v ...interface{}) {
	stdlog.Printf(format, v...)
}

func (l Log) Println(v ...interface{}) {
	stdlog.Println(v...)
}

func (l Log) Fail(exit int, v ...interface{}) {
	stdlog.Println(append([]interface{}{"fail:"}, v...)...)
	os.Exit(exit)
}

func (l Log) Warn(v ...interface{}) {
	stdlog.Println(append([]interface{}{"warn:"}, v...)...)
}

func (l Log) Warnf(format string, v ...interface{}) {
	stdlog.Printf("warn: "+format, v...)
}

func (l Log) Info(v ...interface{}) {
	stdlog.Println(append([]interface{}{"info:"}, v...)...)
}

func (l Log) Infof(format string, v ...interface{}) {
	stdlog.Printf("info: "+format, v...)
}
