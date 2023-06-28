package validator

import (
	"bytes"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	"sort"
)

type Pair struct {
	Key   [gwatCommon.HashLength]byte
	Value gwatCommon.HashArray
}

// processPrevoteData method processes received prevote data to define longest chains of spines which have enough votes
func (vs *Server) processPrevoteData(prevoteData []*ethpb.PreVote, optSpines []gwatCommon.HashArray) []Pair {
	// Divide prevote candidates to subchains of spines and calculate amount of votes
	chains, votes := vs.getChainsAndVotes(prevoteData)

	// Remove invalid subchains
	for k, v := range chains {
		var found bool
		for _, spine := range optSpines {
			if bytes.Contains(spine.ToBytes(), v.ToBytes()) {
				found = true
				break
			}
		}
		if !found {
			delete(chains, k)
			delete(votes, k)
		}
	}

	// Define thresholdVotes and exclude chains that have less votes
	voters := prevoteData[0].GetAggregationBits().Len()
	thresholdVotes := voters/2 + 1
	for k, v := range votes {
		if v < thresholdVotes {
			delete(votes, k)
		}
	}

	// Get a slice of chains sorted by chain length
	sortedByChainLen := vs.sortByChainLen(chains)

	//Define longest chains with enough votes
	bestChains := make([]Pair, 0)
	for _, val := range sortedByChainLen {
		if votes[val.Key] >= thresholdVotes {
			bestChains = append(bestChains, val)
		}
	}
	return bestChains
}

// getChainsAndVotes receives an array of prevote structs, defines unique subchains of spines and calculates total
// amount of votes for these subchains and return data in corresponding maps: map[spinesSubchainHash]spinesSubchain and
// map [spinesSubchainHash]amount of votes
func (vs *Server) getChainsAndVotes(prevote []*ethpb.PreVote) (map[[gwatCommon.HashLength]byte]gwatCommon.HashArray,
	map[[gwatCommon.HashLength]byte]uint64) {
	hashAndChain := make(map[[gwatCommon.HashLength]byte]gwatCommon.HashArray)
	hashAndVotes := make(map[[gwatCommon.HashLength]byte]uint64)

	for _, pv := range prevote {
		can := gwatCommon.HashArrayFromBytes(pv.Data.Candidates)
		for i := 1; i < len(can); i++ {
			chain := can[:i]

			if chain.IsUniq() {
				h := chain.Key()
				hashAndChain[h] = chain
				if helpers.IsAggregatedPrevote(pv) {
					hashAndVotes[h] += pv.GetAggregationBits().Count()
				} else {
					hashAndVotes[h]++
				}
			} else {
				log.Warnf("Prevote spine subchain contains duplicates of hashes for prevote: %v", can)
			}
		}
	}

	return hashAndChain, hashAndVotes
}

// sortByChainLen sorts provided subchains by length and returns as a sorted slice of pairs: subchainHash : subchain
func (vs *Server) sortByChainLen(chainsMap map[[gwatCommon.HashLength]byte]gwatCommon.HashArray) []Pair {
	pairs := make([]Pair, 0, len(chainsMap))
	for key, value := range chainsMap {
		pairs = append(pairs, Pair{Key: key, Value: value})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i].Value) > len(pairs[j].Value)
	})

	return pairs
}
