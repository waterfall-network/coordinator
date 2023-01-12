package operation

import "gitlab.waterfall.network/waterfall/protocol/coordinator/async/event"

// Notifier interface defines the methods of the service that provides beacon block operation updates to consumers.
type Notifier interface {
	OperationFeed() *event.Feed
}
