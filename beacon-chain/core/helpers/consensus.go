package helpers

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

var ErrBadUnpublishedChains = errors.New("bad unpublished chains")

type mapPublications map[gwatCommon.Hash]int

// ConsensusCalcPrefix calculates sequence of prefix from array of unpublished spines sequences.
func ConsensusCalcPrefix(unpublishedChains []gwatCommon.HashArray) (gwatCommon.HashArray, error) {
	if err := ConsensusValidateUnpublishedChains(unpublishedChains); err != nil {
		return gwatCommon.HashArray{}, err
	}
	var (
		publicationsMap = mapPublications{}
		commonChain     = gwatCommon.HashArray{}
		prefix          = gwatCommon.HashArray{}
	)
	for i, chain := range unpublishedChains {
		if i == 0 {
			commonChain = chain
		} else {
			commonChain = commonChain.SequenceIntersection(chain)
		}
		for _, spine := range chain {
			publicationsMap[spine]++
		}
	}

	for _, spine := range commonChain {
		if publicationsMap[spine] >= params.BeaconConfig().SpinePublicationsPefixSupport {
			prefix = append(prefix, spine)
		}
	}

	log.WithFields(log.Fields{
		"prefix":            prefix,
		"unpublishedChains": unpublishedChains,
		"commonChain":       commonChain,
		"publicationsMap":   publicationsMap,
	}).Info("Calculate pefix")

	return prefix, nil
}

// ConsensusValidateUnpublishedChains validate unpublished chains
func ConsensusValidateUnpublishedChains(unpublishedChains []gwatCommon.HashArray) error {
	var firstVal gwatCommon.Hash
	for _, chain := range unpublishedChains {
		// empty chains must be removed
		if len(chain) == 0 {
			return errors.Wrap(ErrBadUnpublishedChains, "contains empty chain")
		}
		// chains must be uniq
		if !chain.IsUniq() {
			return errors.Wrap(ErrBadUnpublishedChains, "chain is not uniq")
		}
		//	the first values of each chain must be equal
		if firstVal == (gwatCommon.Hash{}) {
			firstVal = chain[0]
			continue
		}
		if chain[0] != firstVal {
			return errors.Wrap(ErrBadUnpublishedChains, "the first values of chain are not equal")
		}
	}
	return nil
}

func ConsensusCopyUnpublishedChains(unpublishedChains []gwatCommon.HashArray) []gwatCommon.HashArray {
	cpy := make([]gwatCommon.HashArray, len(unpublishedChains))
	for i, chain := range unpublishedChains {
		cpy[i] = chain.Copy()
	}
	return cpy
}
