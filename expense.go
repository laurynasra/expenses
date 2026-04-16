package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	expenseapi "laurynasra/expenses/api"
	"laurynasra/expenses/internal/engine"
)

func main() {
	supportedProviders := []string{"Wise", "SEB", "Revolut"}
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "tui",
				Usage: "Launch interactive TUI for expense management",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return startTUI()
				},
			},
			{
				Name:  "list-supported",
				Usage: "Lists supported providers",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Supported providers:", supportedProviders)
					return nil
				},
			},
			{
				Name:  "parse-expenses",
				Usage: "Parses expense report for given provider. Prints out aggregated and categorized expenses",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "provider",
						Usage:    "Provider to parse expenses for",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "file",
						Usage:    "File to parse expenses from",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fileName := cmd.String("file")
					provider := cmd.String("provider")
					expenseCategories, parsedExpenses, err := engine.ParseAndAggregate(fileName, provider, "categories.json")
					if err != nil {
						return err
					}
					fmt.Println("Grouped categories:")
					for _, category := range expenseCategories.Categories {
						fmt.Printf("%s: %f\n", category.Category, category.Amount)
						for _, expense := range category.Expenses {
							fmt.Printf("\t%s: %f\n", expense.Description, expense.Amount)
						}
					}
					fmt.Println("Unmatched expenses:")
					for _, expense := range parsedExpenses {
						if !expense.Matched {
							fmt.Printf("%s: %f\n", expense.Description, expense.Amount)
						}
					}
					return nil
				},
			},
			{
				Name:  "serve",
				Usage: "Start HTTP API server for web interface",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "port",
						Usage: "Port to listen on",
						Value: "8080",
					},
					&cli.StringFlag{
						Name:  "categories",
						Usage: "Path to categories JSON file",
						Value: "categories.json",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return expenseapi.StartServer(cmd.String("port"), cmd.String("categories"))
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
