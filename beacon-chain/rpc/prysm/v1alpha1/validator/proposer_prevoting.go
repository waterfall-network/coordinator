package validator

import (
	"bytes"

	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

// processPrevoteData method processes received prevote data to define longest chain of spines which have enough votes
func (vs *Server) processPrevoteData(prevoteData []*ethpb.PreVote, optSpines []gwatCommon.HashArray) gwatCommon.HashArray {
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
			log.WithFields(logrus.Fields{
				"1.thresholdVotes": thresholdVotes,
				"2.votesAmount":    v,
			}).Info("processPrevoteData: chain has less votes than thresholdVotes")
			delete(votes, k)
			delete(chains, k)
		}
	}

	return vs.defineLongestSubchain(chains)
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

// defineLongestSubchain defines and return longest subchain of spines which has amount of votes > votes threshold
func (vs *Server) defineLongestSubchain(chainsMap map[[gwatCommon.HashLength]byte]gwatCommon.HashArray) gwatCommon.HashArray {
	chain := gwatCommon.HashArray{}

	for _, c := range chainsMap {
		if len(c) > len(chain) {
			chain = c
		}
	}

	log.WithFields(logrus.Fields{
		"1.subchain": chain,
	}).Info("defineLongestSubcain: longest subchain with votes > thresholdVotes")

	return chain
}
