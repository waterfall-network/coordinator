package protoarray

import (
	"context"
)

// IsOptimistic returns true if this node is optimistically synced
// A optimistically synced block is synced as usual, but its
// execution payload is not validated, while the EL is still syncing.
// This function returns an error if the block is not found in the fork choice
// store
func (f *ForkChoice) IsOptimistic(root [32]byte) (bool, error) {
	f.store.nodesLock.RLock()
	defer f.store.nodesLock.RUnlock()
	index, ok := f.store.nodesIndices[root]
	if !ok {
		return false, ErrUnknownNodeRoot
	}
	node := f.store.nodes[index]
	return node.status == syncing, nil
}

// SetOptimisticToValid is called with the root of a block that was returned as
// VALID by the EL.
// WARNING: This method returns an error if the root is not found in forkchoice
func (f *ForkChoice) SetOptimisticToValid(ctx context.Context, root [32]byte) error {
	f.store.nodesLock.Lock()
	defer f.store.nodesLock.Unlock()
	// We can only update if given root is in Fork Choice
	index, ok := f.store.nodesIndices[root]
	if !ok {
		return ErrUnknownNodeRoot
	}

	for node := f.store.nodes[index]; node.status == syncing; node = f.store.nodes[index] {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		node.status = valid
		index = node.parent
		if index == NonExistentNode {
			break
		}
	}
	return nil
}
