package dbmarketorders

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/blugelabs/bluge"
	"github.com/epsniff/eveland/src/evesdk"
)

type EveLand interface {
	ListAllMarketOrdersForRegion(ctx context.Context, region *evesdk.Region) ([]*evesdk.MarketOrder, error)
}

type OrderDataDB struct {
	eveSDK EveLand

	dbpath      string
	blugeConfig bluge.Config

	index        *bluge.Writer
	offlineIndex *bluge.OfflineWriter
}

func RemoveDB(dbpath string) error {
	dbdir, err := db_location(dbpath)
	if err != nil {
		return fmt.Errorf("error getting/creating bluge database directory: %v", err)
	}
	// remove the database directory
	err = os.RemoveAll(dbdir)
	if err != nil {
		return fmt.Errorf("error removing bluge database directory: %v", err)
	}
	return nil
}

func New(eveSDK EveLand, dbpath string, isOffline bool) (*OrderDataDB, error) {
	odb := &OrderDataDB{
		eveSDK: eveSDK,
	}

	// Create the database directory if it doesn't exist
	dbdir, err := db_location(dbpath)
	if err != nil {
		return nil, fmt.Errorf("error getting/creating bluge database directory: %v", err)
	}

	odb.dbpath = dbdir
	config := bluge.DefaultConfig(odb.dbpath)
	odb.blugeConfig = config

	if !isOffline {
		w, err := bluge.OpenWriter(config)
		if err != nil {
			return nil, fmt.Errorf("error opening bluge index: %v", err)
		}
		odb.index = w
	}

	if isOffline {
		var offlineIndex, err = bluge.OpenOfflineWriter(config, 1000, 1)
		if err != nil {
			return nil, fmt.Errorf("error opening bluge index writer: %v", err)
		}
		odb.offlineIndex = offlineIndex
	}

	return odb, nil
}

// Close closes the open database.
func (o *OrderDataDB) Close() error {

	if o.index != nil {
		err := o.index.Close()
		if err != nil {
			return fmt.Errorf("error closing Bluge index writer: %v", err)
		}
	}

	if o.offlineIndex != nil {
		err := o.offlineIndex.Close()
		if err != nil {
			err = fmt.Errorf("error closing Bluge index writer: %v", err)
		}
	}

	return nil
}

// GetMarketOrdersBySystemID returns a map of buy orders and a map of sell orders for a given system ID.
// The map keys are the type IDs of the items being sold/bought.
// The map values are MinHeaps and MaxHeaps of the orders for that type ID.
//
// The MinHeap and MaxHeap are sorted by price.
// The SellOrders/MinHeap is sorted in ascending order, so the lowest price is at the top.
// The BuyOrders/MaxHeap is sorted in descending order, so the highest price is at the top.
func (o *OrderDataDB) GetMarketOrdersBySystemID(ctx context.Context, systemID int32) (
	buyOrders map[int32]*MaxHeap, sellOrders map[int32]*MinHeap, err error) {

	if o == nil {
		return nil, nil, fmt.Errorf("OrderDataDB is nil")
	}

	reader, err := o.index.Reader()
	if err != nil {
		return nil, nil, fmt.Errorf("error opening Bluge index reader: %v", err)
	}

	defer func() {
		err = reader.Close()
		if err != nil {
			err = fmt.Errorf("error closing Bluge index reader: %v", err)
		}
	}()

	query := bluge.
		NewNumericRangeQuery(float64(systemID), float64(systemID+1)).
		// NewNumericRangeQuery(0, 1000000000000000000).
		SetField("system_id")
	request := bluge.NewAllMatches(query).WithStandardAggregations()

	// request := bluge.NewAllMatches(bluge.NewMatchAllQuery()).WithStandardAggregations()

	results, err := reader.Search(context.Background(), request)
	if err != nil {
		return nil, nil, fmt.Errorf("error searching index: %v", err)
	}

	buyOrders = map[int32]*MaxHeap{}
	sellOrders = map[int32]*MinHeap{}

	// iterate through the document matches
	match, err := results.Next()
	i := 0
	for err == nil && match != nil {
		// fmt.Printf("Found %v matches, err: %v, match: %v", i, err, match)

		var order *evesdk.MarketOrder = &evesdk.MarketOrder{}
		// load the identifier for this match
		err = match.VisitStoredFields(func(field string, bv []byte) bool {
			value := make([]byte, len(bv))
			copy(value, bv)

			switch field {
			case "_id":
			case "system_id":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.SystemID = int32(tmp)
			case "order_id":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.OrderID = int64(tmp)
			case "type_id":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.TypeID = int32(tmp)
			case "location_id":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.LocationID = int64(tmp)
			case "volume_total":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.VolumeTotal = int32(tmp)
			case "volume_remain":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.VolumeRemain = int32(tmp)
			case "min_volume":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.MinVolume = int32(tmp)
			case "price":
				order.Price, _ = bluge.DecodeNumericFloat64(value)
			case "is_buy_order":
				order.IsBuyOrder, _ = strconv.ParseBool(string(value))
			case "issued":
				order.Issued, _ = bluge.DecodeDateTime(value)
			case "duration":
				tmp, _ := bluge.DecodeNumericFloat64(value)
				order.Duration = int32(tmp)
			case "range":
				order.Range_ = string(value)
			default:
				fmt.Printf("Unknown field: %v\n", field)
			}
			return true
		})
		if err != nil {
			log.Fatalf("error loading stored fields: %v", err)
		}

		if order.IsBuyOrder {
			bos, ok := buyOrders[order.TypeID]
			if !ok {
				bos = NewMaxHeap()
				buyOrders[order.TypeID] = bos
			}
			bos.Push(order)
		} else {
			sos, ok := sellOrders[order.TypeID]
			if !ok {
				sos = NewMinHeap()
				sellOrders[order.TypeID] = sos
			}
			sos.Push(order)
		}

		// load the next document match
		match, err = results.Next()
		i = i + 1
	}
	if err != nil {
		return nil, nil, fmt.Errorf("error iterating through results: %v", err)
	}
	return buyOrders, sellOrders, nil
}

func (o *OrderDataDB) LoadMarketOrders(ctx context.Context, region *evesdk.Region) (found int, err error) {

	// List all market orders.
	if o == nil {
		return 0, fmt.Errorf("OrderDataDB is nil")
	}
	if o.eveSDK == nil {
		return 0, fmt.Errorf("eveSDK is nil")
	}
	orders, err := o.eveSDK.ListAllMarketOrdersForRegion(ctx, region)
	if err != nil {
		return 0, fmt.Errorf("error while trying to list all market orders: %v", err)
	}

	count := 0
	for _, order := range orders {
		orderIdAsBytes := OrderIdKey(order.OrderID)

		//doc := bluge.NewDocument(orderIdAsBytes)
		isBuyOrder := "false"
		if order.IsBuyOrder {
			isBuyOrder = "true"
		}
		doc := bluge.NewDocument(orderIdAsBytes).
			AddField(bluge.NewNumericField("order_id", float64(order.OrderID)).StoreValue()).
			AddField(bluge.NewNumericField("type_id", float64(order.TypeID)).StoreValue()).
			AddField(bluge.NewNumericField("location_id", float64(order.LocationID)).StoreValue()).
			AddField(bluge.NewNumericField("system_id", float64(order.SystemID)).StoreValue()).
			AddField(bluge.NewNumericField("volume_total", float64(order.VolumeTotal)).StoreValue()).
			AddField(bluge.NewNumericField("volume_remain", float64(order.VolumeRemain)).StoreValue()).
			AddField(bluge.NewNumericField("min_volume", float64(order.MinVolume)).StoreValue()).
			AddField(bluge.NewNumericField("price", order.Price).StoreValue()).
			AddField(bluge.NewTextField("is_buy_order", isBuyOrder).StoreValue()).
			AddField(bluge.NewDateTimeField("issued", order.Issued).StoreValue()).
			AddField(bluge.NewNumericField("duration", float64(order.Duration)).StoreValue()).
			AddField(bluge.NewTextField("range", order.Range_).StoreValue())

		o.offlineIndex.Insert(doc)

		count = count + 1
	}

	return count, nil
}

func OrderIdKey(orderId int64) string {
	return strconv.Itoa(int(orderId))
}

func RegionIdKey(regionId int32) string {
	return strconv.Itoa(int(regionId))
}

func db_location(baseDir string) (string, error) {
	dbpath := filepath.Join(baseDir, "orders_bluge_db")

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
