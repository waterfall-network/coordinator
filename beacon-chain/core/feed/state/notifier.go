package state

import "gitlab.waterfall.network/waterfall/protocol/coordinator/async/event"

// Notifier interface defines the methods of the service that provides state updates to consumers.
type Notifier interface {
	StateFeed() *event.Feed
}
