package main

//go:generate go run ./gen/main.go

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/danicc097/todo-ddd-example/internal/generated/client"
)

var (
	debug        bool
	verbose      bool
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	styleDebug   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	styleHeader  = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
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
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show response headers")

	var extraHeaders []string
	rootCmd.PersistentFlags().StringSliceVarP(&extraHeaders, "header", "H", nil, "Custom headers (e.g. -H \"If-None-Match: W/123\")")

	getClient := func() (*client.ClientWithResponses, context.Context) {
		authOption := client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			if token := os.Getenv("API_TOKEN"); token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			for _, h := range extraHeaders {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}

			if debug {
				log.Println(styleDebug.Render("DEBUG: Issuing " + req.Method + " request to " + req.URL.String()))
			}

			return nil
		})

		c, err := client.NewClientWithResponses(apiURL, authOption)
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
