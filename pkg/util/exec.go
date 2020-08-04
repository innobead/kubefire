package util

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

func UpdateDefaultCmdPipes(cmd *exec.Cmd) *exec.Cmd {
	log := NewLogWriter(
		logrus.NewEntry(logrus.StandardLogger()),
		"",
	)

	cmd.Stdout = log
	cmd.Stderr = log
	cmd.Stdin = os.Stdin

	return cmd
}
