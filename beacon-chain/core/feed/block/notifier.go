package block

import "gitlab.waterfall.network/waterfall/protocol/coordinator/async/event"

// Notifier interface defines the methods of the service that provides block updates to consumers.
type Notifier interface {
	BlockFeed() *event.Feed
}
