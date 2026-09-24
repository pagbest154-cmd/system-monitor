package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pagbest154-cmd/system-monitor/internal/hubserver"
	"github.com/pagbest154-cmd/system-monitor/internal/version"
)

func main() {
	mode := flag.String("mode", "hub", "hub or standalone")
	host := flag.String("host", "127.0.0.1", "bind host")
	port := flag.Int("port", 8080, "bind port")
	proxyHeaders := flag.Bool("proxy-headers", false, "trust proxy headers")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version)
		return
	}

	server, err := hubserver.NewServer(*mode)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	server.StartCollector()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	handler := server.Router()
	if *proxyHeaders {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
		})
	}

	httpServer := &http.Server{Addr: addr, Handler: handler}
	go func() {
		log.Printf("system-monitor %s listening on http://%s", version.Version, addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	server.StopCollector()
	_ = httpServer.Close()
}
