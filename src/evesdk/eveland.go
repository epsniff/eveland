package evesdk

import (
	"github.com/antihax/goesi"
)

type eveland struct {
	Eve *goesi.APIClient
}

func New(eve *goesi.APIClient) *eveland {
	return &eveland{Eve: eve}
}
