package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	supportedProviders := []string{"Wise", "SEB", "Revolut"}
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "list-supported",
				Usage: "Lists supported providers",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("Supported providers:", supportedProviders)
					return nil
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
