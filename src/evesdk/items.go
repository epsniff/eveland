package evesdk

import (
	"context"
	"sync"

	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
)

// ListAllTypeIDs returns all types in the game, as an array of TypeIDs.
func (e *eveland) ListAllTypeIDs(ctx context.Context) ([]int32, error) {
	types, resp, err := e.Eve.ESI.UniverseApi.GetUniverseTypes(ctx, nil)
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
	typeMu := &sync.RWMutex{}
	typesAcc := []int32{}

	addTypes := func(myTypes []int32) {
		for _, t := range myTypes {
			typeMu.Lock()
			typesAcc = append(typesAcc, t)
			typeMu.Unlock()
		}
	}
	addTypes(types)

	for i := int32(1); i <= pages; i++ {
		sem <- 1
		go func(page int32) {
			defer func() {
				wg.Done()
				<-sem
			}()
			myTypes, _, err := e.Eve.ESI.UniverseApi.GetUniverseTypes(ctx, &esi.GetUniverseTypesOpts{Page: optional.NewInt32(page)})
			if err != nil {
				return
			}
			addTypes(myTypes)
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish.

	return typesAcc, nil
}

type TypeData struct {
	Capacity    float32 `json:"capacity,omitempty"`    /* capacity number */
	Description string  `json:"description,omitempty"` /* description string */
	// Not sure??   DogmaAttributes []GetUniverseTypesTypeIdDogmaAttribute `json:"dogma_attributes,omitempty"` /* dogma_attributes array */
	// Not sure??   DogmaEffects    []GetUniverseTypesTypeIdDogmaEffect    `json:"dogma_effects,omitempty"`    /* dogma_effects array */
	GraphicId      int32   `json:"graphic_id,omitempty"`      /* graphic_id integer */
	GroupId        int32   `json:"group_id,omitempty"`        /* group_id integer */
	IconId         int32   `json:"icon_id,omitempty"`         /* icon_id integer */
	MarketGroupId  int32   `json:"market_group_id,omitempty"` /* This only exists for types that can be put on the market */
	Mass           float32 `json:"mass,omitempty"`            /* mass number */
	Name           string  `json:"name,omitempty"`            /* name string */
	PackagedVolume float32 `json:"packaged_volume,omitempty"` /* packaged_volume number */
	PortionSize    int32   `json:"portion_size,omitempty"`    /* portion_size integer */
	Published      bool    `json:"published,omitempty"`       /* published boolean */
	Radius         float32 `json:"radius,omitempty"`          /* radius number */
	TypeId         int32   `json:"type_id,omitempty"`         /* type_id integer */
	Volume         float32 `json:"volume,omitempty"`          /* volume number */
}

// GetTypeData returns the type data for a given typeID.
// Use the GetUniverseTypesTypeId endpoint to get the name of a typeID.
func (e *eveland) GetTypeData(ctx context.Context, typeID int32) (*TypeData, error) {
	typeData, _, err := e.Eve.ESI.UniverseApi.GetUniverseTypesTypeId(ctx, typeID, nil)
	if err != nil {
		return nil, err
	}

	t := &TypeData{
		Capacity:    typeData.Capacity,
		Description: typeData.Description,
		// DogmaAttributes: typeData.DogmaAttributes,
		// DogmaEffects:    typeData.DogmaEffects,
		GraphicId:      typeData.GraphicId,
		GroupId:        typeData.GroupId,
		IconId:         typeData.IconId,
		MarketGroupId:  typeData.MarketGroupId,
		Mass:           typeData.Mass,
		Name:           typeData.Name,
		PackagedVolume: typeData.PackagedVolume,
		PortionSize:    typeData.PortionSize,
		Published:      typeData.Published,
		Radius:         typeData.Radius,
		TypeId:         typeData.TypeId,
		Volume:         typeData.Volume,
	}
	// fmt.Println("TypeData:", t)
	return t, nil
}
