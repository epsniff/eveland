package dbmarketorders

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

type OrderDataDB struct {
	pdbs   map[string]*pebble.DB
	eveSDK *evesdk.EveLand

	dbpath string
}

func New(eveSDK *evesdk.EveLand, dbpath string) (*OrderDataDB, error) {
	odb := &OrderDataDB{
		pdbs:   make(map[string]*pebble.DB),
		eveSDK: eveSDK,
	}

	odb.dbpath = dbpath

	return odb, nil

}

// CloseDB closes all open databases.
func (o *OrderDataDB) CloseDB() error {
	for _, pdb := range o.pdbs {
		if err := pdb.Close(); err != nil {
			return fmt.Errorf("error closing db: %v", err)
		}
	}
	return nil
}

// MarketOrdersToMap returns a map of all market orders for each systemID listed.
/*
type MarketOrder struct {
	OrderID      int64     `json:"order_id,omitempty"`
	TypeID       int32     `json:"type_id,omitempty"`
	TypeData     *TypeData `json:"type_data,omitempty"`
	LocationID   int64     `json:"location_id,omitempty"`
	VolumeTotal  int32     `json:"volume_total,omitempty"`
	VolumeRemain int32     `json:"volume_remain,omitempty"`
	MinVolume    int32     `json:"min_volume,omitempty"`
	Price        float64   `json:"price,omitempty"`
	IsBuyOrder   bool      `json:"is_buy_order,omitempty"`
	Issued       time.Time `json:"issued,omitempty"`
	Duration     int32     `json:"duration,omitempty"`
	Range_       string    `json:"range,omitempty"`

	ExpiresIn time.Duration `json:"expires_in,omitempty"`
}
*/
func (o *OrderDataDB) MarketOrdersToMap(systemIDs []int) error {
	panic("not implemented")
}

func (o *OrderDataDB) LoadMarketOrdersFunc(region *evesdk.Region) error {
	err := clear_db_location(o.dbpath, RegionIdKey(region.RegionID))
	if err != nil {
		return fmt.Errorf("error clearing db location: %v", err)
	}

	pebDbPath, err := db_location(o.dbpath, RegionIdKey(region.RegionID))
	if err != nil {
		return fmt.Errorf("error preping db location: %v", err)
	}

	pdb, err := pebble.Open(pebDbPath, &pebble.Options{})
	if err != nil {
		return fmt.Errorf("error opening: %v", err)
	}
	o.pdbs[RegionIdKey(region.RegionID)] = pdb

	// List all market orders.
	orders, err := o.eveSDK.ListAllMarketOrdersForRegion(context.Background(), region)
	if err != nil {
		fmt.Println("Error while trying to list all market orders:", err)
	}

	for _, order := range orders {
		// Write the order details to the db.
		order_json, err := json.Marshal(order)
		if err != nil {
			return fmt.Errorf("error marshalling order: %v", err)
		}
		orderIdAsBytes := OrderIdKey(order.OrderID)

		// fmt.Printf("Writing order %d to db: :%v", order.OrderID, string(order_json))

		err = pdb.Set(orderIdAsBytes, order_json, pebble.Sync)
		if err != nil {
			return fmt.Errorf("error writing to db: %v", err)
		}
	}

	fmt.Printf("Wrote %d orders to db for region %d\n", len(orders), region.RegionID)

	return nil
}

func OrderIdKey(orderId int64) []byte {
	return []byte(strconv.Itoa(int(orderId)))
}

func RegionIdKey(regionId int32) string {
	return strconv.Itoa(int(regionId))
}

func clear_db_location(baseDir, region string) error {
	dbpath := filepath.Join(baseDir, fmt.Sprintf("evemarket_orders_region_%s_peb_db", region))

	err := os.RemoveAll(dbpath)
	if err != nil {
		return fmt.Errorf("could not remove directory %s: %w", dbpath, err)
	}

	return nil
}

func db_location(baseDir, region string) (string, error) {
	dbpath := filepath.Join(baseDir, fmt.Sprintf("evemarket_orders_region_%s_peb_db", region))

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
