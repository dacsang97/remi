// Package main provides the entry point for the Remi application
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dacsang97/remi/internal/app"
)

func main() {
	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nExiting Remi application...")
		os.Exit(0)
	}()

	// Create and run the application
	application := app.NewApplication("remi.yaml")
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
