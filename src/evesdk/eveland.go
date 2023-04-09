package evesdk

import (
	"fmt"

	"github.com/antihax/goesi"
)

type EveLand struct {
	Eve *goesi.APIClient
}

func New(eve *goesi.APIClient) *EveLand {
	return &EveLand{Eve: eve}
}

var ErrNilEveLand = fmt.Errorf("nil eveland")
