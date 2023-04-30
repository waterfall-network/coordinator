package protoarray

import (
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

// GetNode searches root of the first fork of current ForkChoice.
func (f *ForkChoice) GetNode(root [32]byte) *Node {
	if !f.HasNode(root) {
		return nil
	}
	i, _ := f.store.nodesIndices[root]
	return f.store.nodes[i]
}

// GetForks collects forks.
func (f *ForkChoice) GetForks() []*Fork {
	handledRoots := map[[32]byte]*Node{}
	res := []*Fork{}
	for i := f.NodeCount() - 1; i >= 0; i-- {
		node := f.store.nodes[i]
		if _, ok := handledRoots[node.root]; ok {
			continue
		}
		fork := f.GetFork(node.Root())
		if fork == nil {
			//handledRoots[node.Root()] = nil
			continue
		}
		for r, n := range fork.nodesMap {
			handledRoots[r] = n
		}
		res = append(res, fork)
	}
	return res
}

// GetFork collect nodes of tip by recursively iterate by parents.
func (f *ForkChoice) GetFork(root [32]byte) *Fork {
	head := f.GetNode(root)
	if head == nil {
		return nil
	}
	fork := Fork{
		roots:    append(make([][32]byte, 0, f.NodeCount()), root),
		nodesMap: map[[32]byte]*Node{root: head},
	}

	currIndex := head.parent
	for {
		node := f.store.nodes[currIndex]
		if node == nil {
			break
		}
		fork.roots = append(fork.roots, node.Root())
		fork.nodesMap[node.Root()] = node
		if !f.HasParent(node.root) {
			break
		}
		currIndex = node.Parent()
	}
	return &fork
}

// GetCommonAncestor searches the highest common ancestor.
func (f *ForkChoice) GetCommonAncestor() (node *Node) {
	forks := f.GetForks()
	if len(forks) == 0 {
		return nil
	}
	if len(forks) == 1 {
		if len(forks[0].roots) > 0 {
			root := forks[0].roots[0]
			node = forks[0].nodesMap[root]
		}
		return node
	}

	var commonChain gwatCommon.HashArray
	for _, fork := range forks {
		tipRoots := fork.roots
		tipChain := make(gwatCommon.HashArray, len(tipRoots))
		for i, r := range tipRoots {
			tipChain[i] = gwatCommon.BytesToHash(r[:])
		}
		if commonChain == nil {
			commonChain = tipChain.Reverse()
		} else {
			commonChain = commonChain.SequenceIntersection(tipChain.Reverse())
		}
	}
	if len(commonChain) == 0 {
		return nil
	}
	commonRoot := commonChain[len(commonChain)-1]
	return f.GetNode(commonRoot)
}
