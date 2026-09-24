package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pagbest154-cmd/system-monitor/internal/branding"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	icoPath := filepath.Join(root, "packaging", "windows", "app-icon.ico")
	sizes := []int{256, 128, 64, 48, 40, 32, 24, 20, 16}
	if err := branding.SaveICO(icoPath, sizes, "ok"); err != nil {
		fmt.Fprintf(os.Stderr, "gen-icons: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("wrote", icoPath)
}
