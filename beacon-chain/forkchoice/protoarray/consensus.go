package protoarray

import (
	"context"
	"fmt"
	"sort"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"go.opencensus.io/trace"
)

func (f *ForkChoice) setBalances(root [32]byte, balances []uint64) {
	f.store.balancesLock.Lock()
	defer f.store.balancesLock.Unlock()
	f.store.balances[root] = balances
}
func (f *ForkChoice) getBalances(root [32]byte) []uint64 {
	f.store.balancesLock.Lock()
	defer f.store.balancesLock.Unlock()
	return f.store.balances[root]
}

func (f *ForkChoice) setNodeVotes(validator uint64, vote Vote) {
	var (
		maxSlot types.Slot
		index   int
	)
	for i, n := range f.store.nodes {
		if n.slot > maxSlot {
			maxSlot = n.slot
			index = i
		}
	}
	lastNode := f.store.nodes[index]
	lastNode.attsData.votes[validator] = vote
}

func (f *ForkChoice) getNodeVotes(nodeRoot [32]byte) map[uint64]Vote {
	n := f.GetNode(nodeRoot)
	if n == nil {
		return nil
	}
	return n.AttestationsData().Votes()
}

// GetParentByOptimisticSpines retrieves node by root.
func (f *ForkChoice) GetParentByOptimisticSpines(ctx context.Context, optSpines []gwatCommon.HashArray) ([32]byte, error) {
	ctx, span := trace.StartSpan(ctx, "protoArrayForkChoice.GetParentByOptimisticSpines")
	defer span.End()

	f.mu.Lock()
	defer f.mu.Unlock()

	//removes empty values
	_optSpines := make([]gwatCommon.HashArray, 0, len(optSpines))
	for _, ha := range optSpines {
		if len(ha) > 0 {
			_optSpines = append(_optSpines, ha)
		}
	}

	// collect nodes of T(G) tree
	acceptableRootIndexMap, acceptableLeafs := collectTgTreeNodesByOptimisticSpines(f, _optSpines)

	//todo rm
	acceptLeafsArr := make(gwatCommon.HashArray, 0, len(acceptableLeafs))
	for k := range acceptableLeafs {
		acceptLeafsArr = append(acceptLeafsArr, k)
	}
	fRoots := make(gwatCommon.HashArray, len(f.store.nodes))
	for i, n := range f.store.nodes {
		fRoots[i] = n.root
	}
	fc := f
	frk := fc.GetForks()
	frkRoots_0 := gwatCommon.HashArray{}
	if len(frk) > 0 {
		for _, h := range frk[0].roots {
			frkRoots_0 = append(frkRoots_0, h)
		}
	}
	log.WithFields(logrus.Fields{
		"acceptLeafsArr":       len(acceptLeafsArr),
		"fcRoots":              len(fRoots),
		"votes":                len(fc.votes),
		"balances":             len(fc.balances),
		"store.justifiedEpoch": fc.store.justifiedEpoch,
		"store.finalizedEpoch": fc.store.finalizedEpoch,
		//"store.[0].root":       fmt.Sprintf("%#x", fc.store.nodes[0].root),
		//"fork[0]":              frkRoots_0,
	}).Info("Get parent by optimistic spines")

	if len(acceptableRootIndexMap) == 0 {
		return [32]byte{}, nil
	}

	headRoot, err := f.calculateHeadRootByNodesIndexes(ctx, acceptableRootIndexMap)
	if err != nil {
		return [32]byte{}, err
	}

	return headRoot, nil
}

// calculateHeadRootOfForks retrieves root of head of passed forks.
func (f *ForkChoice) calculateHeadRootByNodesIndexes(ctx context.Context, nodesRootIndexMap map[[32]byte]uint64) ([32]byte, error) {

	//todo rm
	fRoots := make(gwatCommon.HashArray, len(f.store.nodes))
	for i, n := range f.store.nodes {
		fRoots[i] = n.root
	}

	frk := f.GetForks()
	frkRoots_0 := gwatCommon.HashArray{}
	if len(frk) > 0 {
		for _, h := range frk[0].roots {
			frkRoots_0 = append(frkRoots_0, h)
		}
	}
	log.WithFields(logrus.Fields{
		"fcRoots":              len(fRoots),
		"votes":                len(f.votes),
		"balances":             len(f.balances),
		"store.justifiedEpoch": f.store.justifiedEpoch,
		"store.finalizedEpoch": f.store.finalizedEpoch,
		//"store.[0].root":       fmt.Sprintf("%#x", f.store.nodes[0].root),
		//"fork[0]":              frkRoots_0,
	}).Info("Calculate head root by nodes indexes")

	// create ForkChoice instance
	fcInstance := New(f.store.justifiedEpoch, f.store.finalizedEpoch)

	// sort node's indexes
	nodeIndexes := make(gwatCommon.SorterAscU64, 0, len(nodesRootIndexMap))
	indexRootMap := make(map[uint64][32]byte, len(nodesRootIndexMap))
	for r, index := range nodesRootIndexMap {
		nodeIndexes = append(nodeIndexes, index)
		indexRootMap[index] = r
	}
	sort.Sort(nodeIndexes)

	// fill ForkChoice instance
	var justifiedRoot [32]byte
	var headRoot [32]byte
	for i, index := range nodeIndexes {

		node := f.store.nodes[index]
		if i == 0 {
			justifiedRoot = node.root
		}
		if fcInstance.HasNode(node.attsData.justifiedRoot) {
			justifiedRoot = node.attsData.justifiedRoot
		}

		n := copyNode(node)
		n.bestChild = NonExistentNode
		n.bestDescendant = NonExistentNode
		n.weight = 0
		if node.parent != NonExistentNode {
			parentRoot, ok := indexRootMap[node.parent]
			if !ok {
				return [32]byte{}, errParentNodFound
			}
			n.parent = fcInstance.store.nodesIndices[parentRoot]
		}
		err := fcInstance.store.insertNode(ctx, n)

		if err != nil {
			return [32]byte{}, err
		}

		// sort validators' indexes
		validatorIndexes := make(gwatCommon.SorterAscU64, 0, len(node.AttestationsData().Votes()))
		for ix := range node.AttestationsData().Votes() {
			validatorIndexes = append(validatorIndexes, ix)
		}
		sort.Sort(validatorIndexes)

		for _, vi := range validatorIndexes {
			vote := n.AttestationsData().votes[vi]
			targetEpoch := vote.nextEpoch
			blockRoot := vote.nextRoot
			// Validator indices will grow the vote cache.
			for vi >= uint64(len(fcInstance.votes)) {
				fcInstance.votes = append(fcInstance.votes, Vote{currentRoot: params.BeaconConfig().ZeroHash, nextRoot: params.BeaconConfig().ZeroHash})
			}

			// Newly allocated vote if the root fields are untouched.
			newVote := fcInstance.votes[vi].nextRoot == params.BeaconConfig().ZeroHash &&
				fcInstance.votes[vi].currentRoot == params.BeaconConfig().ZeroHash

			// Vote gets updated if it's newly allocated or high target epoch.
			if newVote || targetEpoch > fcInstance.votes[vi].nextEpoch {
				fcInstance.votes[vi].nextEpoch = targetEpoch
				fcInstance.votes[vi].nextRoot = blockRoot
			}
		}
	}

	topNode := fcInstance.store.nodes[len(fcInstance.store.nodes)-1]

	// todo check use insead of f.balances
	//balances := f.getBalances(ctx, topNode.root)

	// apply LMD GHOST
	headRoot, err := fcInstance.Head(ctx, topNode.justifiedEpoch, justifiedRoot, f.balances, topNode.finalizedEpoch)

	//todo rm
	log.WithError(err).WithFields(logrus.Fields{
		"headRoot":            fmt.Sprintf("%#x", headRoot),
		"len(nodes)":          len(nodeIndexes),
		"nodeCount":           f.NodeCount(),
		"fc.roots":            len(fcInstance.GetRoots()),
		"_f.roots":            len(f.GetRoots()),
		"topNode.root":        fmt.Sprintf("%#x", topNode.root),
		"parent":              topNode.parent,
		"bestChild":           topNode.bestChild,
		"bestDescendant":      topNode.bestDescendant,
		"weight":              topNode.weight,
		"att.justRoot":        fmt.Sprintf("%#x", topNode.AttestationsData().justifiedRoot),
		"justifiedRoot":       fmt.Sprintf("%#x", justifiedRoot),
		"len(node.att.votes)": len(topNode.AttestationsData().votes),
	}).Info("Get parent by optimistic spines 1")

	if err != nil {
		return [32]byte{}, err
	}

	log.WithFields(logrus.Fields{
		"headRoot":             fmt.Sprintf("%#x", headRoot),
		"votes":                len(fcInstance.votes),
		"_votes":               len(f.votes),
		"balances":             len(fcInstance.balances),
		"_balances":            len(f.balances),
		"store.justifiedEpoch": fcInstance.store.justifiedEpoch,
		"store.finalizedEpoch": fcInstance.store.finalizedEpoch,
		//"store.[0].root":       fmt.Sprintf("%#x", fcInstance.store.nodes[0].root),
	}).Info("Get parent by optimistic spines res")

	return headRoot, nil
}

func collectTgTreeNodesByOptimisticSpines(fc *ForkChoice, optSpines []gwatCommon.HashArray) (map[[32]byte]uint64, map[[32]byte]int) {
	forks := fc.GetForks()
	rootIndexMap := make(map[[32]byte]uint64)
	leafs := make(map[[32]byte]int)
	nodesIndices := fc.store.cpyNodesIndices()

	for frkNr, frk := range forks {
		if frk == nil {
			continue
		}
		for i, r := range frk.roots {
			node := frk.nodesMap[r]

			if len(node.spinesData.cpFinalized) == 0 {
				log.WithFields(logrus.Fields{
					"frkNr":             frkNr,
					"node.index":        i,
					"node.slot":         node.slot,
					"node.root":         fmt.Sprintf("%#x", node.root),
					"node.cpFinalized":  node.spinesData.cpFinalized,
					"node.Finalization": node.spinesData.Finalization(),
					"optSpines":         optSpines,
				}).Error("------ collectTgTreeNodesByOptimisticSpines: checkpoint finalized seq empty ------")
			}

			// rm finalized spines from optSpines if contains
			lastFinHash := node.spinesData.cpFinalized[len(node.spinesData.cpFinalized)-1]
			lastFinIndex := indexOfOptimisticSpines(lastFinHash, optSpines)
			if lastFinIndex > -1 {
				optSpines = optSpines[lastFinIndex+1:]
			}

			// check finalization matches to optSpines
			finalization := node.spinesData.Finalization()
			ok := isSequenceMatchOptimisticSpines(finalization, optSpines)

			log.WithFields(logrus.Fields{
				"ok":           ok,
				"frkNr":        frkNr,
				"node.index":   i,
				"node.slot":    node.slot,
				"node.root":    fmt.Sprintf("%#x", node.root),
				"finalization": len(finalization),
				"optSpines":    len(optSpines),
			}).Info("collectTgTreeNodesByOptimisticSpines: check finalization")

			if !ok {
				continue
			}

			// check prefix matches to optSpines
			prefOptSpines := []gwatCommon.HashArray{}
			if len(optSpines) > len(finalization) {
				prefOptSpines = optSpines[len(finalization):]
			}
			prefix := node.spinesData.Prefix()
			ok = isSequenceMatchOptimisticSpines(prefix, prefOptSpines)

			log.WithFields(logrus.Fields{
				"ok":         ok,
				"frkNr":      frkNr,
				"node.index": i,
				"node.slot":  node.slot,
				"node.root":  fmt.Sprintf("%#x", node.root),
			}).Info("collectTgTreeNodesByOptimisticSpines: check prefix")

			if !ok {
				continue
			}

			//check prefix extension or no published spines
			published := node.spinesData.Spines()
			isExtended := len(prefix.Intersection(published)) > 0 || len(finalization.Intersection(published)) > 0
			if isExtended || len(published) == 0 {
				//collect roots of acceptable forks
				forkRoots := frk.roots[i:]
				for _, root := range forkRoots {
					rootIndexMap[root] = nodesIndices[root]
				}

				log.WithFields(logrus.Fields{
					"ok":         ok,
					"frkNr":      frkNr,
					"node.index": i,
					"node.slot":  node.slot,
					"node.root":  fmt.Sprintf("%#x", node.root),
				}).Info("collectTgTreeNodesByOptimisticSpines: check prefix extension: success")

				leafs[frk.roots[i]] = len(forkRoots)
				break
			}

			// check the first published spine matches to prefOptSpines
			pubOptSpines := []gwatCommon.HashArray{}
			if len(prefOptSpines) > len(prefix) {
				pubOptSpines = prefOptSpines[len(prefix):]
			}

			log.WithFields(logrus.Fields{
				"continue":   len(pubOptSpines) == 0,
				"frkNr":      frkNr,
				"node.index": i,
				"node.slot":  node.slot,
				"node.root":  fmt.Sprintf("%#x", node.root),
			}).Info("collectTgTreeNodesByOptimisticSpines: check published spines length")

			if len(pubOptSpines) == 0 {
				continue
			}

			log.WithFields(logrus.Fields{
				"continue":     !pubOptSpines[0].Has(published[0]),
				"published":    published,
				"pubOptSpines": pubOptSpines,
				"optSpines[0]": optSpines[0],
				"frkNr":        frkNr,
				"node.index":   i,
				"node.slot":    node.slot,
				"node.root":    fmt.Sprintf("%#x", node.root),
			}).Info("collectTgTreeNodesByOptimisticSpines: check the first published spine")

			if !pubOptSpines[0].Has(published[0]) {
				continue
			}
			//collect roots of acceptable forks
			forkRoots := frk.roots[i:]
			for _, root := range forkRoots {
				rootIndexMap[root] = nodesIndices[root]
			}

			log.WithFields(logrus.Fields{
				"frk.roots[i]": fmt.Sprintf("%#x", frk.roots[i]),
				"frkNr":        frkNr,
				"node.index":   i,
				"node.slot":    node.slot,
				"node.root":    fmt.Sprintf("%#x", node.root),
			}).Info("collectTgTreeNodesByOptimisticSpines: check the first published spine: success")

			leafs[frk.roots[i]] = len(forkRoots)
			break
		}
	}
	return rootIndexMap, leafs
}

func isSequenceMatchOptimisticSpines(seq gwatCommon.HashArray, optSpines []gwatCommon.HashArray) bool {
	if len(seq) > len(optSpines) {
		return false
	}
	for i, h := range seq {
		if !optSpines[i].Has(h) {
			return false
		}
	}
	return true
}

func indexOfOptimisticSpines(hash gwatCommon.Hash, optSpines []gwatCommon.HashArray) int {
	for i, sines := range optSpines {
		if sines.Has(hash) {
			return i
		}
	}
	return -1
}

// CollectForkExcludedAttestations collect attestations
func (f *ForkChoice) CollectForkExcludedBlkRoots(leaf gwatCommon.Hash) gwatCommon.HashArray {
	f.mu.Lock()
	defer f.mu.Unlock()
	// collect nodes excluded from fork
	fork := f.GetFork(leaf)
	exIndices := make(map[[32]byte]uint64, len(f.store.nodesIndices))
	for k, v := range f.store.cpyNodesIndices() {
		exIndices[k] = v
	}
	for _, r := range fork.roots {
		delete(exIndices, r)
	}
	//collect attestation
	exRoots := make(gwatCommon.HashArray, 0, len(exIndices))
	for r := range exIndices {
		exRoots = append(exRoots, r)
	}
	return exRoots
}

// insertNode inserts node to the fork choice store's node list.
// It then updates the new node's parent with best child and descendant node.
func (s *Store) insertNode(ctx context.Context, node *Node) error {
	_, span := trace.StartSpan(ctx, "protoArrayForkChoice.insertNode")
	defer span.End()

	s.nodesLock.Lock()
	defer s.nodesLock.Unlock()

	// Return if the block has been inserted into Store before.
	if _, ok := s.nodesIndices[node.root]; ok {
		return nil
	}
	n := copyNode(node)
	index := uint64(len(s.nodes))
	s.nodesIndices[n.root] = index
	s.nodes = append(s.nodes, n)

	// Update parent with the best child and descendant only if it's available.
	if n.parent != NonExistentNode {
		if err := s.updateBestChildAndDescendant(n.parent, index); err != nil {
			return err
		}
	}
	return nil
}

// GetRoots get roots of nodes of forkchoice sorted by indexes.
func (f *ForkChoice) GetRoots() gwatCommon.HashArray {
	nodeIndexes := make(gwatCommon.SorterAscU64, 0, len(f.store.cpyNodesIndices()))
	for _, index := range f.store.cpyNodesIndices() {
		nodeIndexes = append(nodeIndexes, index)
	}
	sort.Sort(nodeIndexes)
	res := make(gwatCommon.HashArray, len(nodeIndexes))
	for i, ix := range nodeIndexes {
		res[i] = gwatCommon.BytesToHash(f.store.nodes[ix].root[:])
	}
	return res
}

// GetNode retrieves node by root.
func (f *ForkChoice) GetNode(root [32]byte) *Node {
	if !f.HasNode(root) {
		return nil
	}
	i := f.store.nodesIndices[root]
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
	if currIndex == NonExistentNode {
		return &fork
	}

	for {
		if currIndex >= uint64(len(f.store.nodes)) {
			log.WithFields(logrus.Fields{
				"currIndex":          currIndex,
				"len(f.store.nodes)": len(f.store.nodes),
			}).Error("FC: get fork oun of range")
			break
		}
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
