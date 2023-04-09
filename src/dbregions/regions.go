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

type EveLand interface {
	ListAllRegions(ctx context.Context) (*evesdk.Regions, error)
}

type RegionsDataDB struct {
	eveSDK EveLand

	pdb *pebble.DB
}

func New(eveSDK EveLand, dbpath string) (*RegionsDataDB, error) {
	pebDbPath, err := db_location(dbpath)
	if err != nil {
		return nil, fmt.Errorf("error prepping db location: %v", err)
	}
	fmt.Println("Storing region data on disk in pebbledb at: ", pebDbPath)

	pdb, err := pebble.Open(pebDbPath, &pebble.Options{})
	if err != nil {
		return nil, fmt.Errorf("error opening: %v", err)
	}

	return &RegionsDataDB{eveSDK: eveSDK, pdb: pdb}, nil
}

func (r *RegionsDataDB) Close() error {
	err := r.pdb.Close()
	if err != nil {
		return fmt.Errorf("error closing: %v", err)
	}

	return nil
}

func (r *RegionsDataDB) LoadRegions(ctx context.Context) error {

	// List all regions.
	regions, err := r.eveSDK.ListAllRegions(context.Background())
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

		err = r.pdb.Set(regionIdAsBytes, region_json, pebble.Sync)
		if err != nil {
			return fmt.Errorf("error writing to db: %v", err)
		}
	}

	return nil
}

// GetRegionByID retrieves a region from the PebbleDB by its regionId.
func (r *RegionsDataDB) GetRegionById(ctx context.Context, regionId int32) (*evesdk.Region, error) {
	regionIdAsBytes := RegionIdKey(regionId)
	value, closer, err := r.pdb.Get(regionIdAsBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading from db: %v", err)
	}
	defer closer.Close()

	var region evesdk.Region
	if err := json.Unmarshal(value, &region); err != nil {
		return nil, fmt.Errorf("error unmarshalling region: %v", err)
	}

	return &region, nil
}

// ListAllRegions retrieves all regions from the PebbleDB.
func (r *RegionsDataDB) ListAllRegions(ctx context.Context) ([]*evesdk.Region, error) {
	var regions []*evesdk.Region
	// Iterate over all regions in the db.
	iter := r.pdb.NewIter(&pebble.IterOptions{})
	defer iter.Close()
	for iter.First(); iter.Valid(); iter.Next() {
		var region evesdk.Region
		if err := json.Unmarshal(iter.Value(), &region); err != nil {
			return nil, fmt.Errorf("error unmarshalling region: %v", err)
		}
		regions = append(regions, &region)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error iterating over regions: %v", err)
	}

	return regions, nil
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
