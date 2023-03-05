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

	eveSDK := evesdk.New(eveClient)

	// List all regions.
	regions, err := eveSDK.ListAllRegions(context.Background())
	if err != nil {
		fmt.Println("Error while trying to list all regions:", err)
	}

	for _, region := range regions.Set {
		fmt.Printf("Region %s has id %d	\n", region.Name, region.RegionID)
	}

	types, err := eveSDK.ListAllTypeIDs(context.Background())
	if err != nil {
		fmt.Println("Error while trying to list all type ids:", err)
	}
	fmt.Printf("Number of types: %v\n", len(types))

	/*
		// sem is a channel that will allow up to 4 concurrent operations.
		var sem = make(chan int, 4)
		// Create a mutex to protect the marketOrders slice
		typeMu := &sync.RWMutex{}
		typeDataList := []*evesdk.TypeData{}
		var wg sync.WaitGroup
		wg.Add(len(types))
		for i, t := range types {
			go func(i int, typeId int32) {
				defer func() {
					wg.Done()
					<-sem
				}()
				td, err := eveSDK.GetTypeData(context.Background(), typeId)
				if err != nil {
					fmt.Println("Error while trying to get type data:", err)
				}
				typeMu.Lock()
				typeDataList = append(typeDataList, td)
				typeMu.Unlock()
				if i%100 == 0 {
					fmt.Printf("Processed %d types.\n", i)
				}
			}(i, t)
		}
		wg.Wait() // Wait for all goroutines to finish.
		fmt.Printf("Number of type data: %v\n", len(typeDataList))
	*/

	/*
		// List all market orders.
		orders, err := evesdk.ListAllMarketOrdersForRegion(context.Background(), regions.Set["Verge Vendor"])
		if err != nil {
			fmt.Println("Error while trying to list all market orders:", err)
		}

		for _, order := range orders {
			fmt.Printf("Order: %s \n", order)
		}
	*/
}
