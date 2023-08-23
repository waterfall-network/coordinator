package validator

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

// processPrevoteData method processes prevote data to define chain which has most votes
func (vs *Server) processPrevoteData(prevoteData []*ethpb.PreVote, optCandidates gwatCommon.HashArray) gwatCommon.HashArray {
	// Divide prevote candidates to subchains of spines
	chains, votes := vs.getChainsAndVotes(prevoteData)

	// Remove invalid subchains
	opc := optCandidates.ToBytes()
	for k, v := range chains {
		if !bytes.Contains(opc, v.ToBytes()) {
			delete(chains, k)
			delete(votes, k)
		}
	}

	return vs.defineMostVotedChain(chains, votes)
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
		for i := 1; i <= len(can); i++ {
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

// defineMostVotedChain defines and returns longest subchain with most of the votes
func (vs *Server) defineMostVotedChain(chainsMap map[[gwatCommon.HashLength]byte]gwatCommon.HashArray,
	votesMap map[[gwatCommon.HashLength]byte]uint64) gwatCommon.HashArray {

	// Define most votes number
	var mostVotes uint64
	for _, v := range votesMap {
		if v > mostVotes {
			mostVotes = v
		}
	}

	// Define subchains which have most of the votes
	chains := make([]gwatCommon.HashArray, 0, len(chainsMap))
	for k, v := range votesMap {
		if v >= mostVotes {
			chains = append(chains, chainsMap[k])
		}
	}

	// Define longest subchain with most of votes
	var chain gwatCommon.HashArray
	for _, v := range chains {
		if len(v) > len(chain) {
			chain = v
		}
	}

	log.WithFields(logrus.Fields{
		"1.votesAmount": mostVotes,
		"2.chain":       chain,
	}).Info("defineMostVotedChain: longest prevote candidates chain with the most votes")
	return chain
}
