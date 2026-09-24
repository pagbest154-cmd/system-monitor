//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	mbOK              = 0x00000000
	mbIconInformation = 0x00000040
	mbIconError       = 0x00000010
)

var (
	user32      = syscall.NewLazyDLL("user32.dll")
	messageBoxW = user32.NewProc("MessageBoxW")
)

func showMessageBox(title, message string, flags uint32) {
	t, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return
	}
	m, err := syscall.UTF16PtrFromString(message)
	if err != nil {
		return
	}
	_, _, _ = messageBoxW.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), uintptr(flags))
}

func showInfo(title, message string) {
	showMessageBox(title, message, mbOK|mbIconInformation)
}

func showError(title, message string) {
	showMessageBox(title, message, mbOK|mbIconError)
}
