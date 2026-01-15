package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudboot/cloudboot-ng/bootos/cb-agent/pkg/agent"
	"github.com/cloudboot/cloudboot-ng/bootos/cb-agent/pkg/client"
)

const (
	AgentVersion = "1.0.0-alpha"
)

func main() {
	// Command-line flags
	serverURL := flag.String("server", getEnv("CB_SERVER_URL", "http://10.0.0.1:8080"), "CloudBoot server URL")
	pollInterval := flag.Duration("poll-interval", 5*time.Second, "Task polling interval")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Print banner
	printBanner()

	// Initialize HTTP client
	httpClient := client.New(*serverURL)
	log.Printf("[INFO] Connecting to CloudBoot server: %s", *serverURL)

	// Create agent
	ag := agent.New(httpClient, agent.Config{
		PollInterval: *pollInterval,
		Debug:        *debug,
	})

	// Run agent
	log.Println("[INFO] Starting BootOS Agent...")
	if err := ag.Run(); err != nil {
		log.Fatalf("[FATAL] Agent failed: %v", err)
	}
}

func printBanner() {
	fmt.Println(`
╔════════════════════════════════════════════╗
║   CloudBoot BootOS Agent                   ║
║   Version: ` + AgentVersion + `                        ║
╚════════════════════════════════════════════╝
`)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
