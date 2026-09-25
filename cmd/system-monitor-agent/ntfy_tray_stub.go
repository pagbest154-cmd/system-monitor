//go:build !windows

package main

func startNtfyListener(configPath string) {}

func registerPushNotifyMenu(func(string)) {}

func currentPushNotifyStatus() string { return "" }
