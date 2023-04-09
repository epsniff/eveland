package dbitems

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/epsniff/eveland/src/evesdk"
)

func LoadItemsFunc(eveSDK *evesdk.EveLand, dbpath string) error {

	pebDbPath, err := db_location(dbpath)
	if err != nil {
		return fmt.Errorf("error preping db location: %v", err)
	}
	fmt.Println("Storing item data on disk in pebbledb at: ", pebDbPath)
	pdb, err := pebble.Open(pebDbPath, &pebble.Options{})
	if err != nil {
		return fmt.Errorf("error opening: %v", err)
	}
	defer func() {
		err := pdb.Close()
		if err != nil {
			fmt.Printf("Error closing db: %v  err: %v \n ", pebDbPath, err)
		}
	}()

	types, err := eveSDK.ListAllTypeIDs(context.Background())
	if err != nil {
		fmt.Println("Error while trying to list all type ids:", err)
	}
	fmt.Printf("Number of types: %v\n", len(types))

	// sem is a channel that will allow up to 4 concurrent operations.
	var sem = make(chan int, 2)
	// Create a mutex to protect the marketOrders slice	append.
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
			try := 0
		retry:
			td, err := eveSDK.GetTypeData(context.Background(), typeId)
			if err != nil {
				err := fmt.Errorf("error while trying to get type data: %v", err)
				fmt.Printf("Error: try: %v, type: %v, err: %v\n", try, typeId, err)
				try++
				time.Sleep(time.Duration(try) * time.Second)
				if try < 5 {
					goto retry
				}
			}
			typeMu.Lock()
			typeDataList = append(typeDataList, td)
			typeMu.Unlock()
			if i%300 == 0 {
				fmt.Printf("Processed %d types.\n", i)
			}
		}(i, t)
	}
	wg.Wait() // Wait for all goroutines to finish.
	fmt.Printf("Number of type data: %v\n", len(typeDataList))

	for _, td := range typeDataList {
		err := pdb.Set(TypeIDKey(td.TypeId), []byte(td.Name), pebble.Sync)
		if err != nil {
			return fmt.Errorf("error while trying to write to db: %v, err: %v", pebDbPath, err)
		}
	}

	fmt.Printf("Done writing to db: %v, number of types: %v\n", pebDbPath, len(typeDataList))

	return nil
}

func TypeIDKey(typeID int32) []byte {
	return []byte(strconv.Itoa(int(typeID)))
}

func db_location(baseDir string) (string, error) {
	dbpath := filepath.Join(baseDir, "eveitems_peb_db")

	_, err := os.Stat(dbpath)
	if os.IsNotExist(err) {
		err := os.Mkdir(dbpath, 0700)
		if err != nil {
			return "", fmt.Errorf("could not create directory %s: %w", dbpath, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("could not stat directory %s: %w", dbpath, err)
	}

	return dbpath, nil
}
