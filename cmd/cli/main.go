package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/danicc097/todo-ddd-example/internal/generated/client"
	"github.com/spf13/cobra"
)

func main() {
	var apiURL string

	rootCmd := &cobra.Command{
		Use:   "todo-cli",
		Short: "CLI for Todo API",
	}
	rootCmd.PersistentFlags().StringVar(&apiURL, "url", "http://127.0.0.1:8090/api/v1", "API Server URL")

	getClient := func() (*client.ClientWithResponses, context.Context) {
		c, err := client.NewClientWithResponses(apiURL)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		return c, context.Background()
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all todos",
		Run: func(cmd *cobra.Command, args []string) {
			c, ctx := getClient()

			resp, err := c.GetAllTodosWithResponse(ctx)
			if err != nil {
				log.Fatalf("Request failed: %v", err)
			}

			if resp.JSON200 == nil {
				log.Fatalf("Error: status %d", resp.StatusCode())
			}

			for _, t := range *resp.JSON200 {
				fmt.Printf("[%s] %s (%s)\n", t.Id, t.Title, t.Status)
			}
		},
	}

	createCmd := &cobra.Command{
		Use:   "create [title]",
		Short: "Create a new todo",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c, ctx := getClient()
			title := args[0]

			body := client.CreateTodoRequest{Title: title}

			resp, err := c.CreateTodoWithResponse(ctx, &client.CreateTodoParams{}, body)
			if err != nil {
				log.Fatalf("Request failed: %v", err)
			}

			if resp.JSON201 == nil {
				log.Fatalf("Failed to create: status %d", resp.StatusCode())
			}

			fmt.Printf("Created Todo ID: %s\n", resp.JSON201.Id)
		},
	}

	rootCmd.AddCommand(listCmd, createCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
