package cmd

import (
	"context"
	"fmt"

	"github.com/epsniff/eveland/src/evesdedb"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

func addSDEUtilsCommands(rootCmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {
	evesde, err := evesdedb.New(dbpath)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	var SDEUtilsCmd = &cobra.Command{
		Use:   "sdetables",
		Short: "sdetables",
		Run: func(cmd *cobra.Command, args []string) {
			if err := evesde.ShowAllTables(context.TODO()); err != nil {
				fmt.Println("error: ", err)
			}
		},
	}

	var sdecmdTables = &cobra.Command{
		Use:   "sdeshowjumpscols",
		Short: "sdeshowjumpscols",
	}
	table := "mapSolarSystemJumps"
	// add a flag to pick the table
	sdecmdTables.PersistentFlags().StringVarP(&table, "table", "t", "mapSolarSystemJumps", `table to show columns for
	go run main.go sdeshowjumpscols    
	# the default is mapSolarSystemJumps
	    go run main.go sdeshowjumpscols  -t=mapSolarSystemJumps
		go run main.go sdeshowjumpscols  -t=mapDenormalize
		go run main.go sdeshowjumpscols  -t=mapSolarSystems
	`)
	sdecmdTables.Run = func(cmd *cobra.Command, args []string) {
		// TODO - make this a flag to determin which table
		if err := evesde.ShowAllColumns(context.TODO(), table); err != nil {
			fmt.Println("error: ", err)
		}
	}

	rootCmd.AddCommand(SDEUtilsCmd)
	rootCmd.AddCommand(sdecmdTables)
}
