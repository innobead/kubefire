package util

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

func UpdateCommandDefaultLogWithInfo(cmd *exec.Cmd) *exec.Cmd {
	return UpdateCommandDefaultLog(cmd, logrus.InfoLevel)
}

func UpdateCommandDefaultLog(cmd *exec.Cmd, logLevel logrus.Level) *exec.Cmd {
	log := NewLogWriter(
		logrus.NewEntry(logrus.StandardLogger()),
		logLevel,
		"",
	)

	cmd.Stdout = log
	cmd.Stderr = log
	cmd.Stdin = os.Stdin

	return cmd
}
