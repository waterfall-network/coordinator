package genesis

import (
	_ "embed"
	"fmt"

	"github.com/golang/snappy"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
)

var (
	//go:embed mainnet.ssz.snappy
	mainnetRawSSZCompressed []byte // 1.8Mb
	//go:embed testnet8.ssz.snappy
	testnet8RawSSZCompressed []byte // 1.8Mb
)

// State returns a copy of the genesis state from a hardcoded value.
func State(netName, genesisPath string) (state.BeaconState, error) {
	switch netName {
	case params.ConfigNames[params.Mainnet]:
		if genesisPath == "" {
			// todo activate load mainnetRawSSZCompressed after implemented
			/*depth
			  "//io/file:go_default_library",
			  "@com_github_status_im_keycard_go//hexutils:go_default_library",
			*/
			//serState, err := file.ReadFileAsBytes("<path-to>/coordinator-genesis.ssz")
			//if err != nil {
			//	return nil, err
			//}
			//encodedState := snappy.Encode(nil, serState)
			//err = file.WriteFile("<path-to>/testnet8.ssz.snappy", []byte(encodedState))
			//if err != nil {
			//	return nil, err
			//}
			if false {
				return load(mainnetRawSSZCompressed)
			}
			return nil, fmt.Errorf("mainnet raw genesis is not installed. Use cmd param `--genesis-state` to define path to genesis.ssz")
		}
		return nil, nil
	case params.ConfigNames[params.Testnet8]:
		if genesisPath == "" {
			return load(testnet8RawSSZCompressed)
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
	b, err := snappy.Decode(nil /*dst*/, b)
	if err != nil {
		return nil, err
	}
	if err = st.UnmarshalSSZ(b); err != nil {
		return nil, err
	}
	return v1.InitializeFromProtoUnsafe(st)
}
