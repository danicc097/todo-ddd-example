package main

//go:generate go run ./gen/main.go

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/danicc097/todo-ddd-example/internal/generated/client"
)

func main() {
	var apiURL string

	defaultURL := "http://127.0.0.1:8090/api/v1"

	if envURL := os.Getenv("API_URL"); envURL != "" {
		if strings.Contains(envURL, "/api/") {
			defaultURL = envURL
		} else {
			defaultURL = strings.TrimRight(envURL, "/") + "/api/v1"
		}
	}

	rootCmd := &cobra.Command{
		Use:   "todo-cli",
		Short: "CLI for Todo API",
	}

	rootCmd.PersistentFlags().StringVar(&apiURL, "url", defaultURL, "API Server URL (also set via API_URL env var)")

	getClient := func() (*client.ClientWithResponses, context.Context) {
		c, err := client.NewClientWithResponses(apiURL)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		return c, context.Background()
	}

	RegisterGeneratedCommands(rootCmd, getClient)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
