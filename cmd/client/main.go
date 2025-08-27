package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kryptx/internal/config"
	"kryptx/internal/gui"
	"kryptx/internal/network"
	"kryptx/internal/utils"
)

var (
	configPath = flag.String("config", "configs/client.yaml", "Config file path")
	guiMode    = flag.Bool("gui", true, "Run with GUI")
	verbose    = flag.Bool("v", false, "Verbose logging")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := utils.NewLogger(*verbose)
	logger.Info("Starting KryptX VPN Client")

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize VPN client
	vpnClient, err := network.NewVPNClient(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize VPN client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Info("Shutting down...")
		cancel()
		vpnClient.Disconnect()
		os.Exit(0)
	}()

	if *guiMode {
		// Start GUI
		app := gui.NewApp(vpnClient, cfg, logger)
		app.Run()
	} else {
		// CLI mode
		fmt.Println("KryptX VPN Client - CLI Mode")
		if err := vpnClient.Connect(ctx); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}

		// Keep running until signal
		<-ctx.Done()
	}
}
