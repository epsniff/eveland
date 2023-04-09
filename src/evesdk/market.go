package evesdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
)

type MarketOrder struct {
	OrderID      int64     `json:"order_id,omitempty"`
	TypeID       int32     `json:"type_id,omitempty"`
	TypeData     *TypeData `json:"type_data,omitempty"`
	LocationID   int64     `json:"location_id,omitempty"`
	SystemID     int32     `json:"system_id,omitempty"`
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

func (m *MarketOrder) String() string {
	s, err := json.Marshal(m)
	if err != nil {
		return "json.Marshal failed: " + err.Error()
	}
	return string(s)
}

// ListAllMarketOrdersForRegion returns all market orders for a region.
// ToDo: Add regionID as parameter.
func (e *EveLand) ListAllMarketOrdersForRegion(ctx context.Context, region *Region) ([]*MarketOrder, error) {

	const allOrderType = "all" // ToDo: Add orderType as parameter.
	orders, resp, err := e.Eve.ESI.MarketApi.GetMarketsRegionIdOrders(ctx, allOrderType, region.RegionID, nil)
	if err != nil {
		return nil, err
	}

	// Extract the number of pages from the response header	and use it to get the other pages concurrently.
	pages, err := getPages(resp)
	if err != nil {
		return nil, err
	}

	// Create a waitgroup to keep track of the goroutines
	var wg sync.WaitGroup
	wg.Add(int(pages))
	// sem is a channel that will allow up to 4 concurrent operations.
	var sem = make(chan int, 4)

	// Create a mutex to protect the marketOrders slice
	marketMu := &sync.RWMutex{}
	marketOrders := []*MarketOrder{}

	typeDataCache := make(map[int32]*TypeData)

	addOrders := func(orders []esi.GetMarketsRegionIdOrders200Ok) {
		for _, order := range orders {
			m := &MarketOrder{
				OrderID:      order.OrderId,
				TypeID:       order.TypeId,
				LocationID:   order.LocationId,
				SystemID:     order.SystemId,
				VolumeTotal:  order.VolumeTotal,
				VolumeRemain: order.VolumeRemain,
				MinVolume:    order.MinVolume,
				Price:        order.Price,
				IsBuyOrder:   order.IsBuyOrder,
				Issued:       order.Issued,
				Duration:     order.Duration,
				Range_:       order.Range_,
				ExpiresIn:    timeUntilCacheExpires(resp),
			}

			marketMu.RLock()
			typeData, ok := typeDataCache[m.TypeID]
			marketMu.RUnlock()

			if !ok {
				marketMu.Lock()
				// Check again in case another goroutine already got it.
				typeData, ok = typeDataCache[m.TypeID]
				if !ok {
					typeData, err = e.GetTypeData(ctx, m.TypeID)
					if err != nil {
						fmt.Println("GetType failed:", err)
					} else {
						typeDataCache[m.TypeID] = typeData
					}
				}
				marketMu.Unlock()
			}
			m.TypeData = typeData

			marketMu.Lock()
			marketOrders = append(marketOrders, m)
			marketMu.Unlock()
		}
	}
	addOrders(orders)

	// Get the other pages concurrently. We skip page 0 because we already have it.
	for i := 1; int32(i) <= pages; i++ {
		sem <- 1
		go func(page int32) {
			defer func() {
				<-sem
				wg.Done()
			}()

			orders, _, err = e.Eve.ESI.MarketApi.GetMarketsRegionIdOrders(
				ctx,
				allOrderType,
				region.RegionID,
				&esi.GetMarketsRegionIdOrdersOpts{Page: optional.NewInt32(page)},
			)
			addOrders(orders)
		}(int32(i))
	}

	wg.Wait() // Wait for everything to finish

	return marketOrders, nil
}
