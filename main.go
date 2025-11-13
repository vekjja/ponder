/*
Copyright © 2023 Kevin.Jayne@iCloud.com
*/
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/vekjja/ponder/cmd"
)

func main() {
	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle signals in a goroutine
	go func() {
		<-sigChan
		cmd.CleanupAndExit()
		os.Exit(0)
	}()

	cmd.Execute()
}
