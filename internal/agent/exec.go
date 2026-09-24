package agent

import "os/exec"

func execCommand(cmd string, args []string) error {
	c := exec.Command(cmd, args...)
	return c.Start()
}
