package util

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"reflect"
	"strings"
)

type LogWriter struct {
	Log      *logrus.Entry
	logLevel logrus.Level
	prefix   string

	logf  func(format string, args ...interface{})
	logln func(args ...interface{})
}

func NewLogWriter(log *logrus.Entry, level logrus.Level, prefix string) io.Writer {
	writer := &LogWriter{Log: log, logLevel: level, prefix: prefix}
	writerValue := reflect.ValueOf(writer).Elem().FieldByName("Log")

	if prefix != "" {
		funcName := strings.Title(fmt.Sprintf("%sf", level.String()))
		writer.logf = writerValue.MethodByName(funcName).Interface().(func(format string, args ...interface{}))
	} else {
		funcName := strings.Title(fmt.Sprintf("%sln", level.String()))
		writer.logln = writerValue.MethodByName(funcName).Interface().(func(args ...interface{}))
	}

	return writer
}

func (l LogWriter) Write(p []byte) (n int, err error) {
	size := len(p)
	if size == 0 {
		return 0, nil
	}

	if l.prefix != "" {
		l.logf("%s: %s", l.prefix, string(p))
	} else {
		l.logln(string(p))
	}

	return size, nil
}
