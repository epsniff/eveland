package dbregions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/pebble"
	"github.com/epsniff/eveland/src/evesdk"
)

func LoadRegionsFunc(eveSDK *evesdk.EveLand, dbpath string) error {

	pebDbPath, err := db_location(dbpath)
	if err != nil {
		return fmt.Errorf("error preping db location: %v", err)
	}
	fmt.Println("Storing region data on disk in pebbledb at: ", pebDbPath)
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

	// List all regions.
	regions, err := eveSDK.ListAllRegions(context.Background())
	if err != nil {
		return fmt.Errorf("error listing regions: %v", err)
	}

	for _, region := range regions.Set {
		// Write the region details to the db.
		region_json, err := json.Marshal(region)
		if err != nil {
			return fmt.Errorf("error marshalling region: %v", err)
		}
		regionIdAsBytes := RegionIdKey(region.RegionID)

		// fmt.Printf("Writing region %d to db: :%v", region.RegionID, string(region_json))

		err = pdb.Set(regionIdAsBytes, region_json, pebble.Sync)
		if err != nil {
			return fmt.Errorf("error writing to db: %v", err)
		}
	}

	return nil
}

func RegionIdKey(regionId int32) []byte {
	return []byte(strconv.Itoa(int(regionId)))
}

func db_location(baseDir string) (string, error) {
	dbpath := filepath.Join(baseDir, "everegions_peb_db")

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
