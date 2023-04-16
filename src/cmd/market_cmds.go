package cmd

import (
	"context"
	"fmt"

	"github.com/epsniff/eveland/src/dbmarketorders"
	"github.com/epsniff/eveland/src/dbregions"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

func addMarketOrdersCommands(rootCmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {
	// eveland loadmarketorders
	var LoadMarketOrdersCmd = &cobra.Command{
		Use:   "loadmarketorders",
		Short: "loadmarketorders",
		// go run main.go  loadmarketorders
		Run: func(cmd *cobra.Command, args []string) {
			dbmarketorders.RemoveDB(dbpath)

			dbm, err := dbmarketorders.New(eveSDK, dbpath)
			if err != nil {
				fmt.Println("error creating db marketorders: ", err)
				return
			}

			coreTradeRegions := map[string]struct{}{
				"Aridia":       struct{}{},
				"Black Rise":   struct{}{},
				"Derelik":      struct{}{},
				"Devoid":       struct{}{},
				"Domain":       struct{}{},
				"Essence":      struct{}{},
				"Everyshore":   struct{}{},
				"Genesis":      struct{}{},
				"Heimatar":     struct{}{},
				"Kador":        struct{}{},
				"Lonetrek":     struct{}{},
				"Metropolis":   struct{}{},
				"Molden Heath": struct{}{},
				"Oasa":         struct{}{},
				"Placid":       struct{}{},
				"Pochven":      struct{}{},
				"Solitude":     struct{}{},
				"Sinq Laison":  struct{}{},
				"Syndicate":    struct{}{},
				"Tash-Murkon":  struct{}{},
				"The Citadel":  struct{}{},
				"The Forge":    struct{}{},
				"Verge Vendor": struct{}{},
			}

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

			// Iterate through the regions and load market data
			for _, region := range regions {
				if _, ok := coreTradeRegions[region.Name]; !ok {
					continue
				}
				cnt, err := dbm.LoadMarketOrders(context.TODO(), region)
				if err != nil {
					fmt.Printf("error loading market orders for region %s: %v", region.Name, err)
					return
				}
				fmt.Printf("Loaded %d market orders for region %s \n", cnt, region.Name)
			}

			if err := dbm.Close(); err != nil {
				fmt.Println("error: ", err)
				return
			}
		},
	}

	rootCmd.AddCommand(LoadMarketOrdersCmd)
}
