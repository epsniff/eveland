package cmd

import (
	"context"
	"fmt"

	"github.com/epsniff/eveland/src/dbregions"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

func addRegionCommands(rootCmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {
	// eveland loadregions
	var LoadRegionsCmd = &cobra.Command{
		Use:   "loadregions",
		Short: "loadregions",
		Run: func(cmd *cobra.Command, args []string) {
			dbr, err := dbregions.New(eveSDK, dbpath)
			if err != nil {
				fmt.Println("error creating db regions: ", err)
				return
			}
			if err := dbr.LoadRegions(context.Background()); err != nil {
				fmt.Println("error: ", err)
			}

			// print all regions.
			regions, err := dbr.ListAllRegions(context.Background())
			if err != nil {
				fmt.Println("error listing regions: ", err)
				return
			}
			for _, region := range regions {
				fmt.Printf("Region: name: %s, id: %d\n", region.Name, region.RegionID)
			}
		},
	}
	rootCmd.AddCommand(LoadRegionsCmd)
}
