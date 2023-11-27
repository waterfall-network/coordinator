package p2p

import (
	"context"

	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/enode"
)

// filterNodes wraps an iterator such that Next only returns nodes for which
// the 'check' function returns true. This custom implementation also
// checks for context deadlines so that in the event the parent context has
// expired, we do exit from the search rather than  perform more network
// lookups for additional peers.
func filterNodes(ctx context.Context, it enode.Iterator, check func(*enode.Node) bool) enode.Iterator {
	return &filterIter{ctx, it, check}
}

type filterIter struct {
	context.Context
	enode.Iterator
	check func(*enode.Node) bool
}

// Next looks up for the next valid node according to our
// filter criteria.
// https://github.com/golangci/golangci-lint/discussions/2287
// nolint: typecheck
func (f *filterIter) Next() bool {
	seen := make(map[enode.ID]struct{})
	for f.Iterator.Next() {
		if f.Context.Err() != nil {
			return false
		}
		pNode := f.Node()
		if _, ok := seen[pNode.ID()]; ok {
			return false
		}
		seen[pNode.ID()] = struct{}{}
		if f.check(pNode) {
			return true
		}
	}
	return false
}
