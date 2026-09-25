package agent

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/pagbest154-cmd/system-monitor/internal/paths"
)

var (
	logMu     sync.Mutex
	logFile   *os.File
	stdLogger *log.Logger
)

func InitLogging() {
	logMu.Lock()
	defer logMu.Unlock()
	if stdLogger != nil {
		return
	}
	_ = os.MkdirAll(paths.ConfigDir, 0o755)
	path := filepath.Join(paths.ConfigDir, "agent.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		stdLogger = log.New(os.Stderr, "system-monitor-agent: ", log.LstdFlags)
		return
	}
	logFile = f
	stdLogger = log.New(f, "system-monitor-agent: ", log.LstdFlags)
}

func AgentLogf(format string, args ...interface{}) {
	logMu.Lock()
	logger := stdLogger
	logMu.Unlock()
	if logger == nil {
		InitLogging()
		logMu.Lock()
		logger = stdLogger
		logMu.Unlock()
	}
	if logger == nil {
		return
	}
	logger.Printf(format, args...)
}

func WrapRequestError(op, endpoint string, err error) error {
	if err == nil {
		return nil
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return fmt.Errorf("%s %s: DNS lookup failed for %q (%s)", op, endpoint, dns.Name, dns.Err)
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return fmt.Errorf("%s %s: network %s failed: %v", op, endpoint, opErr.Op, opErr.Err)
	}
	return fmt.Errorf("%s %s: %w", op, endpoint, err)
}

func WrapHTTPError(op, endpoint string, status int, body string) error {
	snippet := body
	if len(snippet) > 240 {
		snippet = snippet[:240] + "..."
	}
	if snippet == "" {
		return fmt.Errorf("%s %s: HTTP %d", op, endpoint, status)
	}
	return fmt.Errorf("%s %s: HTTP %d: %s", op, endpoint, status, snippet)
}
