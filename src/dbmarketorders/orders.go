package dbmarketorders

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/epsniff/eveland/src/evesdk"
)

type OrderDataDB struct {
	indexWriter *bluge.Writer
	indexReader *bluge.Reader
	eveSDK      *evesdk.EveLand

	dbpath string
}

func New(eveSDK *evesdk.EveLand, dbpath string) (*OrderDataDB, error) {
	odb := &OrderDataDB{
		eveSDK: eveSDK,
	}

	odb.dbpath = dbpath
	config := bluge.DefaultConfig(odb.dbpath)

	w, err := bluge.OpenWriter(config)
	if err != nil {
		return nil, fmt.Errorf("error opening Bluge index: %v", err)
	}
	odb.indexWriter = w

	r, err := w.Reader()
	if err != nil {
		return nil, fmt.Errorf("error opening Bluge index reader: %v", err)
	}
	odb.indexReader = r

	return odb, nil

}

// CloseDB closes the open database.
func (o *OrderDataDB) CloseDB() error {
	err := o.indexWriter.Close()
	if err != nil {
		return fmt.Errorf("error closing Bluge index writer: %v", err)
	}
	err = o.indexReader.Close()
	if err != nil {
		return fmt.Errorf("error closing Bluge index reader: %v", err)
	}
	return nil
}

func (o *OrderDataDB) GetMarketOrdersBySystemID(systemID int32) ([]evesdk.MarketOrder, error) {
	termQuery := bluge.NewTermQuery(strconv.Itoa(int(systemID)))
	termQuery.SetField("system_id")

	request := bluge.NewTopNSearch(10, termQuery)
	results, err := o.indexReader.Search(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("error searching index: %v", err)
	}

	marketOrders := map[int64]*evesdk.MarketOrder{}

	// iterate through the document matches
	match, err := results.Next()
	for err == nil && match != nil {

		var order *evesdk.MarketOrder
		// load the identifier for this match
		err = match.VisitStoredFields(func(field string, value []byte) bool {
			if field == "_id" {
				fmt.Printf("match: %s\n", string(value))
			}
			return true
		})
		if err != nil {
			log.Fatalf("error loading stored fields: %v", err)
		}

		marketOrders[order.OrderID] = order

		// load the next document match
		match, err = results.Next()
	}
	if err != nil {
		return nil, fmt.Errorf("error iterating through results: %v", err)
	}
	return marketOrders, nil
}

func (o *OrderDataDB) LoadMarketOrdersFunc(region *evesdk.Region) error {
	// List all market orders.
	orders, err := o.eveSDK.ListAllMarketOrdersForRegion(context.Background(), region)
	if err != nil {
		fmt.Println("Error while trying to list all market orders:", err)
	}

	for oIdx, order := range orders {
		orderIdAsBytes := OrderIdKey(order.OrderID)

		doc := bluge.NewDocument(orderIdAsBytes)

		orderValue := reflect.ValueOf(order)
		orderType := orderValue.Type()
		batch := bluge.NewBatch()

		for i := 0; i < orderValue.NumField(); i++ {
			fieldValue := orderValue.Field(i)
			fieldType := orderType.Field(i)

			jsonTag := fieldType.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			switch fieldValue.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				doc.AddField(bluge.NewNumericField(jsonTag, float64(fieldValue.Int())).StoreValue())
			case reflect.Float32, reflect.Float64:
				doc.AddField(bluge.NewNumericField(jsonTag, fieldValue.Float()).StoreValue())
			case reflect.Bool:
				val := "false"
				if fieldValue.Bool() {
					val = "true"
				}
				doc.AddField(bluge.NewTextField(jsonTag, val).StoreValue())
			case reflect.String:
				doc.AddField(bluge.NewTextField(jsonTag, fieldValue.String()).StoreValue())
			}

			if fieldType.Type == reflect.TypeOf(time.Time{}) {
				doc.AddField(bluge.NewDateTimeField(jsonTag, fieldValue.Interface().(time.Time)).StoreValue())
			}
		}

		batch.Update(doc.ID(), doc)
		if oIdx%1000 == 0 {
			err = o.indexWriter.Batch(batch)
			if err != nil {
				return fmt.Errorf("error writing to index: %v", err)
			}
			batch = bluge.NewBatch()
		}
	}

	fmt.Printf("Wrote %d orders to index for region %d\n", len(orders), region.RegionID)

	return nil
}

func OrderIdKey(orderId int64) string {
	return strconv.Itoa(int(orderId))
}

func RegionIdKey(regionId int32) string {
	return strconv.Itoa(int(regionId))
}
