package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmundt/stitch/sample/server"
)

func main() {
	handler := server.NewHandler()
	addr := "localhost:8080"
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("stitch demo listening on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("demo server failed: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Print("shutting down demo server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed, forcing close: %v", err)
		_ = httpServer.Close()
	}
	log.Print("demo server stopped")
}
