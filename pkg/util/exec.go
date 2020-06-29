package util

import (
	"os"
	"os/exec"
)

func UpdateDefaultCmdPipes(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
}
