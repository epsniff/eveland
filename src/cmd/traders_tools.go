package cmd

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/epsniff/eveland/src/dbitems"
	"github.com/epsniff/eveland/src/dbmarketorders"
	"github.com/epsniff/eveland/src/evesdedb"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/spf13/cobra"
)

func addTradersToolsCommands(rootCmd *cobra.Command, eveSDK *evesdk.EveLand, dbpath string) {

	var systemName = "Scheenins"
	var jumps = 3
	var salesTax = 0.036 // 3.6% sales tax with accounting level 5, use (1 - salesTax) to convert, e.g. .036 to .964
	var maxCargoSize = 16_000.0
	var minProfit = 1_000_000

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

			dbi, err := dbitems.New(eveSDK, dbpath)
			if err != nil {
				fmt.Println("error creating db items: ", err)
				return
			}

			bestRevenueOrders := make(map[int32]*dbmarketorders.MaxHeap)
			bestAcquires := make(map[int32]*dbmarketorders.MinHeap)

			systems := 0
			ts := 0
			ta := 0
			for _, system := range systemsInRange {
				systems++
				traderRevenues, traderAcquires, err := dbm.GetMarketOrdersBySystemID(context.TODO(), int32(system.ID))
				if err != nil {
					fmt.Println("error getting market orders: ", err)
					return
				}
				if traderRevenues == nil || traderAcquires == nil {
					fmt.Println("The system doesn't have any buy/sells, for system: ", system.ID)
					continue
				}
				for typeID, tSells := range traderRevenues {
					if _, ok := bestRevenueOrders[typeID]; !ok {
						bestRevenueOrders[typeID] = dbmarketorders.NewMaxHeap()
					}
					bestRevenueOrders[typeID].Merge(tSells)
					ts += tSells.Cnt()
				}
				for typeID, acquires := range traderAcquires {
					if _, ok := bestAcquires[typeID]; !ok {
						bestAcquires[typeID] = dbmarketorders.NewMinHeap()
					}
					bestAcquires[typeID].Merge(acquires)
					ta += acquires.Cnt()
				}
			}

			items := []map[string]interface{}{}
			for typeID, revenueOpportunity := range bestRevenueOrders {
				// Given an opportunity to sell, find the best acquiring price
				acquireMinHeap, ok := bestAcquires[typeID]
				if !ok {
					// no one is selling this item
					continue
				}
				bestAcquireOption := acquireMinHeap.Peek()
				bestRevenueOpportunity := revenueOpportunity.Peek()

				td, err := dbi.GetItem(context.TODO(), typeID)
				if err != nil {
					fmt.Println("error getting item: ", err)
					return
				}

				maxCargo := math.Floor(maxCargoSize / float64(td.Volume))

				tradersAcquireCost := bestAcquireOption.Price * float64(bestAcquireOption.VolumeRemain)
				quantity := math.Min(float64(bestRevenueOpportunity.VolumeRemain), float64(bestAcquireOption.VolumeRemain))
				quantity = math.Min(quantity, maxCargo)
				tradersSellRevenue := bestRevenueOpportunity.Price * quantity * (1 - salesTax) // convert .036 to .964

				cargoVolume := quantity * float64(td.Volume)

				profit := int(tradersSellRevenue - tradersAcquireCost)
				if profit < minProfit {
					// fmt.Println("skipping name: ", td.Name, "profit: ", profit)
					continue
				}

				revSystemName, err := evesde.SystemIDToName(bestRevenueOpportunity.SystemID)
				if err != nil {
					fmt.Println("error getting system name: ", err)
					return
				}
				acquireSystemName, err := evesde.SystemIDToName(bestAcquireOption.SystemID)
				if err != nil {
					fmt.Println("error getting system name: ", err)
					return
				}

				jumps, err := evesde.ShortestPath(bestRevenueOpportunity.SystemID, bestAcquireOption.SystemID)
				if err != nil {
					fmt.Println("error getting shortest path: ", err)
					return
				}

				item := map[string]interface{}{
					"type_id": typeID,
					"name":    td.Name,
					"buy_from": map[string]interface{}{
						"price":      int(bestAcquireOption.Price),
						"vol_remain": bestAcquireOption.VolumeRemain,
						"system":     acquireSystemName,
						"system_id":  bestAcquireOption.SystemID,
					},
					"sell_to": map[string]interface{}{
						"price":      int(bestRevenueOpportunity.Price),
						"vol_remain": bestRevenueOpportunity.VolumeRemain,
						"system":     revSystemName,
						"system_id":  bestRevenueOpportunity.SystemID,
					},
					"cargo_volume":     cargoVolume,
					"item_size":        td.Volume,
					"quantity":         quantity,
					"profit_after_tax": profit,
					"jumps":            jumps,
				}
				items = append(items, item)
			}

			// sort items by profit
			//sort.Slice(items, func(i, j int) bool {
			//	return items[i]["profit_after_tax"].(int) > items[j]["profit_after_tax"].(int)
			//})

			// sort items by profit per jump
			sort.Slice(items, func(i, j int) bool {
				ji := items[i]["jumps"].([]int)
				jj := items[j]["jumps"].([]int)
				return len(ji) < len(jj)
			})

			// print out the top 10 items
			cnt := math.Min(100, float64(len(items)))
			fmt.Println("top 10 items: ")
			for i := 0; i < int(cnt); i++ {
				it := items[i]
				fmt.Printf("   %d: profit_after_tax: %d name: [%s] quantity: %v item_size: %v cargo_volume: %v  link: https://evetycoon.com/market/%d  \n",
					i+1, it["profit_after_tax"], it["name"], it["quantity"], it["item_size"], it["cargo_volume"], it["type_id"])
				fmt.Printf("      trader acquire details : price: %d vol_remain: %v system: %s\n",
					it["buy_from"].(map[string]interface{})["price"],
					it["buy_from"].(map[string]interface{})["vol_remain"],
					it["buy_from"].(map[string]interface{})["system"])
				fmt.Printf(
					"      trader sell details    : price: %d vol_remain: %v system: %s\n",
					it["sell_to"].(map[string]interface{})["price"],
					it["sell_to"].(map[string]interface{})["vol_remain"],
					it["sell_to"].(map[string]interface{})["system"])

				path := []string{}
				jumps := it["jumps"].([]int)
				for _, j := range jumps {
					n, err := evesde.SystemIDToName(int32(j))
					if err != nil {
						fmt.Println("error getting system name {path}: ", err)
						return
					}
					path = append(path, n)
				}
				fmt.Printf("      jumps: %v shortest path: %v\n", len(jumps), path)
				fmt.Printf("        https://everoute.net/index.php?startSystemName=%v&endSystemName=%v&comparisonType=safer \n",
					it["buy_from"].(map[string]interface{})["system"].(string),
					it["sell_to"].(map[string]interface{})["system"].(string))
			}
			fmt.Println("")

		},
	}
	FindBestTradeRouteCmd.PersistentFlags().
		StringVarP(&systemName, "system", "s", "Scheenins", "system name to use as the center of the search. default is Scheenins.")
	FindBestTradeRouteCmd.PersistentFlags().
		IntVarP(&jumps, "jumps", "j", 3, "number of jumps to search. default is 3.")

	rootCmd.AddCommand(FindBestTradeRouteCmd)
}
