package testing

import (
	"github.com/waterfall-foundation/coordinator/time/slots"
)

var _ slots.Ticker = (*MockTicker)(nil)
