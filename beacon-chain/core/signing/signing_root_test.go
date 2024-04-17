package signing_test

import (
	"bytes"
	"context"
	"testing"

	fuzz "github.com/google/gofuzz"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/signing"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/crypto/bls"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/util"
)

func TestSigningRoot_ComputeSigningRoot(t *testing.T) {
	emptyBlock := util.NewBeaconBlock()
	_, err := signing.ComputeSigningRoot(emptyBlock, bytesutil.PadTo([]byte{'T', 'E', 'S', 'T'}, 32))
	assert.NoError(t, err, "Could not compute signing root of block")
}

func TestSigningRoot_ComputeDomain(t *testing.T) {
	tests := []struct {
		epoch      uint64
		domainType [4]byte
		domain     []byte
	}{
		{epoch: 1, domainType: [4]byte{4, 0, 0, 0}, domain: []byte{4, 0, 0, 0, 245, 165, 253, 66, 209, 106, 32, 48, 39, 152, 239, 110, 211, 9, 151, 155, 67, 0, 61, 35, 32, 217, 240, 232, 234, 152, 49, 169}},
		{epoch: 2, domainType: [4]byte{4, 0, 0, 0}, domain: []byte{4, 0, 0, 0, 245, 165, 253, 66, 209, 106, 32, 48, 39, 152, 239, 110, 211, 9, 151, 155, 67, 0, 61, 35, 32, 217, 240, 232, 234, 152, 49, 169}},
		{epoch: 2, domainType: [4]byte{5, 0, 0, 0}, domain: []byte{5, 0, 0, 0, 245, 165, 253, 66, 209, 106, 32, 48, 39, 152, 239, 110, 211, 9, 151, 155, 67, 0, 61, 35, 32, 217, 240, 232, 234, 152, 49, 169}},
		{epoch: 3, domainType: [4]byte{4, 0, 0, 0}, domain: []byte{4, 0, 0, 0, 245, 165, 253, 66, 209, 106, 32, 48, 39, 152, 239, 110, 211, 9, 151, 155, 67, 0, 61, 35, 32, 217, 240, 232, 234, 152, 49, 169}},
		{epoch: 3, domainType: [4]byte{5, 0, 0, 0}, domain: []byte{5, 0, 0, 0, 245, 165, 253, 66, 209, 106, 32, 48, 39, 152, 239, 110, 211, 9, 151, 155, 67, 0, 61, 35, 32, 217, 240, 232, 234, 152, 49, 169}},
	}
	for _, tt := range tests {
		if got, err := signing.ComputeDomain(tt.domainType, nil, nil); !bytes.Equal(got, tt.domain) {
			t.Errorf("wanted domain version: %d, got: %d", tt.domain, got)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestSigningRoot_ComputeDomainAndSign(t *testing.T) {
	tests := []struct {
		name       string
		genState   func(t *testing.T) (state.BeaconState, []bls.SecretKey)
		genBlock   func(t *testing.T, st state.BeaconState, keys []bls.SecretKey) *ethpb.SignedBeaconBlock
		domainType [4]byte
		want       []byte
	}{
		{
			name: "block proposer",
			genState: func(t *testing.T) (state.BeaconState, []bls.SecretKey) {
				beaconState, privKeys := util.DeterministicGenesisState(t, 100)
				require.NoError(t, beaconState.SetSlot(beaconState.Slot()+1))
				return beaconState, privKeys
			},
			genBlock: func(t *testing.T, st state.BeaconState, keys []bls.SecretKey) *ethpb.SignedBeaconBlock {
				block, err := util.GenerateFullBlock(st, keys, &util.BlockGenConfig{NumDeposits: 0}, 1)
				require.NoError(t, err)
				return block
			},
			domainType: params.BeaconConfig().DomainBeaconProposer,
			want: []byte{
				0x90, 0x52, 0xb9, 0xca, 0x12, 0x39, 0xe4, 0x20, 0x3, 0x81, 0x42, 0xbe, 0x12, 0xb2, 0x1f, 0xa3,
				0x61, 0xde, 0x45, 0xa, 0xbf, 0xb5, 0xf5, 0x2b, 0x97, 0x6d, 0x8c, 0x26, 0x93, 0x11, 0x71,
				0xff, 0x3c, 0x58, 0x2, 0x20, 0x2b, 0x73, 0xe8, 0xe9, 0x1e, 0x68, 0xfd, 0x33, 0xbf, 0x2f, 0xd3,
				0x5d, 0xa, 0xa, 0xc1, 0xf2, 0xcd, 0xaa, 0x39, 0x23, 0xdc, 0xfd, 0x6a, 0xdf, 0xb8, 0xa1, 0x7,
				0x30, 0xdd, 0xba, 0x34, 0xb5, 0xc0, 0x41, 0x40, 0x2c, 0xe9, 0x1d, 0xf1, 0xb4, 0x29, 0xbd, 0xcd,
				0xd7, 0x4f, 0x6d, 0xb7, 0x77, 0x71, 0x22, 0x10, 0x58, 0xfb, 0xa3, 0xc4, 0xae, 0xd4, 0x12, 0xba, 0xcc,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beaconState, privKeys := tt.genState(t)
			idx, err := helpers.BeaconProposerIndex(context.Background(), beaconState)
			require.NoError(t, err)
			block := tt.genBlock(t, beaconState, privKeys)
			got, err := signing.ComputeDomainAndSign(
				beaconState, time.CurrentEpoch(beaconState), block, tt.domainType, privKeys[idx])
			require.NoError(t, err)
			require.DeepEqual(t, tt.want, got, "Incorrect signature")
		})
	}
}

func TestSigningRoot_ComputeForkDigest(t *testing.T) {
	tests := []struct {
		version []byte
		root    [32]byte
		result  [4]byte
	}{
		{version: []byte{'A', 'B', 'C', 'D'}, root: [32]byte{'i', 'o', 'p'}, result: [4]byte{0x69, 0x5c, 0x26, 0x47}},
		{version: []byte{'i', 'm', 'n', 'a'}, root: [32]byte{'z', 'a', 'b'}, result: [4]byte{0x1c, 0x38, 0x84, 0x58}},
		{version: []byte{'b', 'w', 'r', 't'}, root: [32]byte{'r', 'd', 'c'}, result: [4]byte{0x83, 0x34, 0x38, 0x88}},
	}
	for _, tt := range tests {
		digest, err := signing.ComputeForkDigest(tt.version, tt.root[:])
		require.NoError(t, err)
		assert.Equal(t, tt.result, digest, "Wanted domain version: %#x, got: %#x", digest, tt.result)
	}
}

func TestFuzzverifySigningRoot_10000(_ *testing.T) {
	fuzzer := fuzz.NewWithSeed(0)
	state := &ethpb.BeaconState{}
	pubkey := [fieldparams.BLSPubkeyLength]byte{}
	sig := [96]byte{}
	domain := [4]byte{}
	var p []byte
	var s []byte
	var d []byte
	for i := 0; i < 10000; i++ {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(&pubkey)
		fuzzer.Fuzz(&sig)
		fuzzer.Fuzz(&domain)
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(&p)
		fuzzer.Fuzz(&s)
		fuzzer.Fuzz(&d)
		err := signing.VerifySigningRoot(state, pubkey[:], sig[:], domain[:])
		_ = err
		err = signing.VerifySigningRoot(state, p, s, d)
		_ = err
	}
}
