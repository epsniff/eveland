package cmd

import (
	"context"
	"fmt"

	"github.com/epsniff/eveland/src/dbitems"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

func addItemCommands(rootCmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {
	// eveland loaditems
	var LoadItemsCmd = &cobra.Command{
		Use:   "loaditems",
		Short: "loaditems",
		Run: func(cmd *cobra.Command, args []string) {
			dbi, err := dbitems.New(eveSDK, dbpath)
			if err != nil {
				fmt.Println("error creating db items: ", err)
				return
			}
			if err := dbi.LoadItems(context.Background()); err != nil {
				fmt.Println("error: ", err)
			}
			if err := dbi.Close(); err != nil {
				fmt.Println("error: ", err)
				return
			}
		},
	}

	rootCmd.AddCommand(LoadItemsCmd)
}
