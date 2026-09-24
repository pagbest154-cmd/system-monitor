package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pagbest154-cmd/system-monitor/internal/agent"
	"github.com/pagbest154-cmd/system-monitor/internal/paths"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

func main() {
	configPath := flag.String("config", paths.AgentConfig, "path to agent.yaml")
	tray := flag.Bool("tray", false, "run system tray (Windows)")
	settings := flag.Bool("settings", false, "open settings UI (Windows)")
	checkUpdate := flag.Bool("check-update", false, "check for updates")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version)
		return
	}

	if *checkUpdate {
		result := agent.CheckForUpdates(true)
		fmt.Println(agent.FormatUpdateMessage(result))
		if result.Error != nil {
			fmt.Println(*result.Error)
			os.Exit(1)
		}
		if result.UpdateAvailable {
			os.Exit(2)
		}
		os.Exit(0)
	}

	if *tray {
		if err := runTray(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "tray: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if *settings {
		if err := runSettings(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "settings: %v\n", err)
			os.Exit(1)
		}
		return
	}

	runner, err := agent.NewRunner(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent: %v\n", err)
		os.Exit(1)
	}
	runner.Start()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	runner.Stop()
}
