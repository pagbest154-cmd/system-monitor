//go:build !windows

package hiddenexec

import "os/exec"

func Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
