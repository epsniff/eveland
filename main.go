package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/antihax/goesi"
	"github.com/epsniff/eveland/src/evecache"
	"github.com/epsniff/eveland/src/evesdk"
	"github.com/gregjones/httpcache"
)

func main() {
	// Create our HTTP Client with a http a cache transport.
	c, err := evecache.New()
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

	evesdk := evesdk.New(eveClient)

	// List all regions.
	regions, err := evesdk.ListAllRegions(context.Background())
	if err != nil {
		fmt.Println("Error while trying to list all regions:", err)
	}

	for _, region := range regions.Set {
		fmt.Printf("Region %s has id %d	\n", region.Name, region.RegionID)
	}

	// List all market orders.
	_, err = evesdk.ListAllMarketOrdersForRegion(context.Background(), regions.Set["Verge Vendor"])
	if err != nil {
		fmt.Println("Error while trying to list all market orders:", err)
	}
	//for _, order := range orders {
	//	fmt.Printf("Order: %s \n", order)
	//}
}
