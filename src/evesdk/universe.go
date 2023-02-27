package evesdk

import "context"

type Region struct {
	RegionID    int32  `json:"region_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Regions struct {
	Set map[string]*Region `json:"regions,omitempty"`
}

func (e *eveland) ListAllRegions(ctx context.Context) (*Regions, error) {
	regionsIDs, _, err := e.Eve.ESI.UniverseApi.GetUniverseRegions(ctx, nil)
	if err != nil {
		return nil, err
	}

	regions := &Regions{Set: make(map[string]*Region)}

	for _, region := range regionsIDs {
		regionInfo, _, err := e.Eve.ESI.UniverseApi.GetUniverseRegionsRegionId(ctx, region, nil)
		if err != nil {
			return nil, err
		}
		r := &Region{RegionID: region, Name: regionInfo.Name, Description: regionInfo.Description}
		regions.Set[regionInfo.Name] = r
	}

	return regions, nil
}
