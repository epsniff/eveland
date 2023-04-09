package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/antihax/goesi"
	"github.com/epsniff/eveland/src/dbitems"
	"github.com/epsniff/eveland/src/dbjumpcal"
	"github.com/epsniff/eveland/src/dbmarketorders"
	"github.com/epsniff/eveland/src/dbregions"
	"github.com/epsniff/eveland/src/dbsdeutils"
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

func Init(eveSDK *evesdk.EveLand, dbpath string) {
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

	var LoadMarketOrdersCmd = &cobra.Command{
		Use:   "loadmarketorders",
		Short: "loadmarketorders",
		Run: func(cmd *cobra.Command, args []string) {
			dbm, err := dbmarketorders.New(eveSDK, dbpath)
			if err != nil {
				fmt.Println("error creating db marketorders: ", err)
				return
			}

			// TODO - make this a flag and lookup the region id from the db.
			reg_verge_vendor := &evesdk.Region{
				RegionID: 10000068,
				Name:     "Verge Vendor",
			}

			if cnt, err := dbm.LoadMarketOrders(context.TODO(), reg_verge_vendor); err != nil {
				fmt.Printf("error loading market orders for region %s: %v", reg_verge_vendor.Name, err)
				return
			} else {
				fmt.Printf("Loaded %d market orders for region %s \n", cnt, reg_verge_vendor.Name)
			}

			reg_sinq_laison := &evesdk.Region{
				RegionID: 10000032,
				Name:     "Sinq Laison",
			}
			if cnt, err := dbm.LoadMarketOrders(context.TODO(), reg_sinq_laison); err != nil {
				fmt.Printf("error loading market orders for region %s: %v", reg_sinq_laison.Name, err)
				return
			} else {
				fmt.Printf("Loaded %d market orders for region %s \n", cnt, reg_sinq_laison.Name)
			}

			if err := dbm.Close(); err != nil {
				fmt.Println("error: ", err)
				return
			}
		},
	}
	rootCmd.AddCommand(LoadMarketOrdersCmd)

	var LoadItemDataCmd = &cobra.Command{
		Use:   "loaditems",
		Short: "loaditems",
		Run: func(cmd *cobra.Command, args []string) {
			if err := dbitems.LoadItemsFunc(eveSDK, dbpath); err != nil {
				fmt.Println("error: ", err)
			}
		},
	}
	rootCmd.AddCommand(LoadItemDataCmd)

	var SDEUtilsCmd = &cobra.Command{
		Use:   "sdetables",
		Short: "sdetables",
		Run: func(cmd *cobra.Command, args []string) {
			if err := dbsdeutils.ShowAllTables(dbpath); err != nil {
				fmt.Println("error: ", err)
			}
		},
	}
	rootCmd.AddCommand(SDEUtilsCmd)

	var SDEUtilsCmd2 = &cobra.Command{
		Use:   "sdeshowjumpscols",
		Short: "sdeshowjumpscols",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO - make this a flag to determin which table
			if err := dbsdeutils.ShowAllColumns(dbpath, "mapSolarSystemJumps"); err != nil {
				fmt.Println("error: ", err)
			}
			//if err := dbsdeutils.ShowAllColumns(dbpath, "mapDenormalize"); err != nil {
			//	fmt.Println("error: ", err)
			//}
			//if err := dbsdeutils.ShowAllColumns(dbpath, "mapSolarSystems"); err != nil {
			//	fmt.Println("error: ", err)
			//}
		},
	}
	rootCmd.AddCommand(SDEUtilsCmd2)

	var FindBestTradeRouteCmd = &cobra.Command{
		Use:   "best-trade",
		Short: "best-trade",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO - make these flags and look up for the
			var systemName = "Odebeinn"
			var jumps = 5

			if sysId, err := dbjumpcal.GetSystemID(systemName, dbpath); err != nil {
				fmt.Println("error: ", err)
			} else if res, err := dbjumpcal.SystemsWithinNJumps(sysId, jumps, dbpath); err != nil {
				fmt.Println("error: ", err)
			} else {
				fmt.Printf("Systems within %v jumps of %v:  results:\n%v\n", 5, systemName, res)
			}
		},
	}
	rootCmd.AddCommand(FindBestTradeRouteCmd)

	var JumpDistanceCmd = &cobra.Command{
		Use:   "jumps",
		Short: "jumps",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO - make these flags and look up for the
			var systemName = "Odebeinn"

			if sysId, err := dbjumpcal.GetSystemID(systemName, dbpath); err != nil {
				fmt.Println("error: ", err)
			} else if res, err := dbjumpcal.SystemsWithinNJumps(sysId, 5, dbpath); err != nil {
				fmt.Println("error: ", err)
			} else {
				fmt.Printf("Systems within %v jumps of %v:  results:\n%v\n", 5, systemName, res)
			}
		},
	}
	rootCmd.AddCommand(JumpDistanceCmd)

	var GetSystemIDFromNameCmd = &cobra.Command{
		Use:   "system-id",
		Short: "system-id",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO - make this a flag and lookup the region id from the db.
			var systemName = "Odebeinn"
			if sysId, err := dbjumpcal.GetSystemID(systemName, dbpath); err != nil {
				fmt.Println("error: ", err)
			} else {
				fmt.Printf("SystemName: %v System ID: %v\n", systemName, sysId)
			}
		},
	}
	rootCmd.AddCommand(GetSystemIDFromNameCmd)
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

	// Get our ESI (EVE API) API Client.
	eveClient := goesi.NewAPIClient(client, "early testing, contact esniff@gmail.com")

	eveSDK := evesdk.New(eveClient)

	// Initialize our CLI.
	Init(eveSDK, os.Getenv("GOPATH")+"/src/github.com/epsniff/eveland/_data")
	Execute()
}
