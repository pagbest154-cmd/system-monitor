//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pagbest154-cmd/system-monitor/internal/hiddenexec"
	"golang.org/x/sys/windows"
)

func launchSettings(configPath string) error {
	exe, err := os.Executable()
	if err != nil {
		showError("system-monitor agent", "Не удалось определить путь к программе:\n"+err.Error())
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		exe, _ = os.Executable()
	}

	params := []string{"--settings"}
	if configPath != "" {
		params = append(params, "--config", configPath)
	}

	verb, _ := windows.UTF16PtrFromString("open")
	exePtr, _ := windows.UTF16PtrFromString(exe)
	paramPtr, _ := windows.UTF16PtrFromString(strings.Join(params, " "))
	dirPtr, _ := windows.UTF16PtrFromString(filepath.Dir(exe))

	if err := windows.ShellExecute(0, verb, exePtr, paramPtr, dirPtr, windows.SW_SHOW); err != nil {
		// Fallback: cmd start (works when ShellExecute is blocked).
		startArgs := []string{"/C", "start", "", quoteWindows(exe)}
		startArgs = append(startArgs, params...)
		cmd := hiddenexec.Command("cmd.exe", startArgs...)
		cmd.Dir = filepath.Dir(exe)
		if startErr := cmd.Start(); startErr != nil {
			msg := fmt.Sprintf("Не удалось открыть настройки:\n%v\n%v", err, startErr)
			showError("system-monitor agent", msg)
			return fmt.Errorf("%s", msg)
		}
	}
	return nil
}

func quoteWindows(s string) string {
	if strings.ContainsAny(s, " \t") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}
