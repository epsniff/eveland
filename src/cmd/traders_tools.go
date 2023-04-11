package cmd

import (
	"context"
	"fmt"

	"github.com/epsniff/eveland/src/dbmarketorders"
	"github.com/epsniff/eveland/src/evesdedb"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

func addTradersToolsCommands(rootCmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {

	var systemName = "Odebeinn"
	var jumps = 5
	var FindBestTradeRouteCmd = &cobra.Command{
		Use:   "best-trades",
		Short: "best-trades",
		Long: ` TODO FILL THIS IN
`,
		Run: func(cmd *cobra.Command, args []string) {
			evesde, err := evesdedb.New(dbpath)
			if err != nil {
				fmt.Println("error: ", err)
				return
			}
			sysId, err := evesde.GetSystemID(systemName)
			if err != nil {
				fmt.Println("error: ", err)
				return
			}
			systemsInRange, err := evesde.SystemsWithinNJumps(sysId, jumps)
			if err != nil {
				fmt.Println("error: ", err)
				return
			}

			dbm, err := dbmarketorders.New(eveSDK, dbpath)
			if err != nil {
				fmt.Println("error creating db marketorders: ", err)
				return
			}

			/*
				dbi, err := dbitems.New(eveSDK, dbpath)
				if err != nil {
					fmt.Println("error creating db items: ", err)
					return
				}
			*/

			// bestBuyOrders := make(map[string]*evesdk.MarketOrder)
			// bestSellOrders := make(map[string]*evesdk.MarketOrder)
			for _, system := range systemsInRange {
				buys, sells, err := dbm.GetMarketOrdersBySystemID(context.TODO(), int32(system.ID))
				if err != nil {
					fmt.Println("error getting market orders: ", err)
					return
				}
				if buys == nil {
					fmt.Println("buys is nil, for system: ", system.ID)
				}
				if sells == nil {
					fmt.Println("sells is nil, for system: ", system.ID)
				}
			}
		},
	}
	FindBestTradeRouteCmd.PersistentFlags().
		StringVarP(&systemName, "system", "s", "Odebeinn", "system name to use as the center of the search. default is Odebeinn.")
	FindBestTradeRouteCmd.PersistentFlags().
		IntVarP(&jumps, "jumps", "j", 5, "number of jumps to search. default is 5.")

	rootCmd.AddCommand(FindBestTradeRouteCmd)
}
