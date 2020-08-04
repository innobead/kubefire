package util

import (
	"github.com/sirupsen/logrus"
	"io"
)

type LogWriter struct {
	log    *logrus.Entry
	prefix string
}

func NewLogWriter(log *logrus.Entry, prefix string) io.Writer {
	return &LogWriter{log: log, prefix: prefix}
}

func (l LogWriter) Write(p []byte) (n int, err error) {
	size := len(p)
	if size == 0 {
		return 0, nil
	}

	if l.prefix != "" {
		l.log.Infof("%s: %s", l.prefix, string(p))
	} else {
		l.log.Infoln(string(p))
	}

	return size, nil
}
