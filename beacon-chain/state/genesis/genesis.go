package genesis

import (
	_ "embed"
	"fmt"

	"github.com/waterfall-foundation/coordinator/beacon-chain/state"
	v1 "github.com/waterfall-foundation/coordinator/beacon-chain/state/v1"
	"github.com/waterfall-foundation/coordinator/config/params"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
)

var (
	//go:embed mainnet.ssz.snappy
	mainnetRawSSZCompressed []byte // 1.8Mb
)

// State returns a copy of the genesis state from a hardcoded value.
func State(netName, genesisPath string) (state.BeaconState, error) {
	switch netName {
	case params.ConfigNames[params.Mainnet]:
		if genesisPath == "" {
			// todo activate load mainnetRawSSZCompressed after implemented
			if false {
				return load(mainnetRawSSZCompressed)
			}
			return nil, fmt.Errorf("mainnet raw genesis is not installed. Use cmd param `--genesis-state` to define path to genesis.ssz")
		}
		return nil, nil
	default:
		// No state found.
		return nil, nil
	}
}

// load a compressed ssz state file into a beacon state struct.
func load(b []byte) (state.BeaconState, error) {
	st := &ethpb.BeaconState{}
	//b, err := snappy.Decode(nil /*dst*/, b)
	//if err != nil {
	//	return nil, err
	//}
	if err := st.UnmarshalSSZ(b); err != nil {
		return nil, err
	}
	return v1.InitializeFromProtoUnsafe(st)
}
