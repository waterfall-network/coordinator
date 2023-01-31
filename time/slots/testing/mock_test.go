package testing

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
)

var _ slots.Ticker = (*MockTicker)(nil)
