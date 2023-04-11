package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/antihax/goesi"
	"github.com/epsniff/eveland/src/cmd"
	"github.com/epsniff/eveland/src/evecache"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/gregjones/httpcache"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eve",
	Short: "eve",
	Long:  `eve`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No command specified. Please specify a command.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func main() {

	storagePath := os.Getenv("GOPATH") + "/src/github.com/epsniff/eveland/_data"

	fmt.Printf("using the following path for database files: %v\n", storagePath)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		fmt.Println("error: ", err)
		return
	}

	// Create our HTTP Client with a http a cache transport.
	c, err := evecache.New(storagePath)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err := c.Close()
		if err != nil {
			fmt.Println("Error while trying to close evecache:", err)
		}
	}()
	tp := httpcache.NewTransport(c)
	client := &http.Client{Transport: tp}

	// Get our ESI (EVE API) API Client with our custom caching transport.
	eveClient := goesi.NewAPIClient(client, "early testing, contact esniff@gmail.com")
	eveSDK := evesdk.New(eveClient)

	// Initialize our CLI.
	cmd.Register(rootCmd, eveSDK, storagePath)

	// Execute our CLI.
	Execute()
}
