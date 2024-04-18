package util

import (
	"bytes"
	"context"
	"encoding/hex"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
	"google.golang.org/protobuf/proto"
)

func TestSetupInitialDeposits_1024Entries(t *testing.T) {
	entries := 1
	resetCache()
	deposits, privKeys, err := DeterministicDepositsAndKeys(uint64(entries))
	require.NoError(t, err)
	_, depositDataRoots, err := DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)

	if len(deposits) != entries {
		t.Fatalf("incorrect number of deposits returned, wanted %d but received %d", entries, len(deposits))
	}
	if len(privKeys) != entries {
		t.Fatalf("incorrect number of private keys returned, wanted %d but received %d", entries, len(privKeys))
	}
	expectedPublicKeyAt0 := []byte{0xa9, 0x9a, 0x76, 0xed, 0x77, 0x96, 0xf7, 0xbe, 0x22, 0xd5, 0xb7, 0xe8, 0x5d, 0xee, 0xb7, 0xc5, 0x67, 0x7e, 0x88, 0xe5, 0x11, 0xe0, 0xb3, 0x37, 0x61, 0x8f, 0x8c, 0x4e, 0xb6, 0x13, 0x49, 0xb4, 0xbf, 0x2d, 0x15, 0x3f, 0x64, 0x9f, 0x7b, 0x53, 0x35, 0x9f, 0xe8, 0xb9, 0x4a, 0x38, 0xe4, 0x4c}
	if !bytes.Equal(deposits[0].Data.PublicKey, expectedPublicKeyAt0) {
		t.Fatalf("incorrect public key, wanted %x but received %x", expectedPublicKeyAt0, deposits[0].Data.PublicKey)
	}
	expectedWithdrawalCredentialsAt0 := []byte{0x54, 0x7D, 0xCB, 0xA5, 0xBA, 0xC1, 0x6A, 0x89, 0x10, 0x8B, 0x6B, 0x6A, 0x1F, 0xE3, 0x69, 0x5D, 0x1A, 0x87, 0x4A, 0xB}
	if !bytes.Equal(deposits[0].Data.WithdrawalCredentials, expectedWithdrawalCredentialsAt0) {
		t.Fatalf("incorrect withdrawal credentials, wanted %x but received %x", expectedWithdrawalCredentialsAt0, deposits[0].Data.WithdrawalCredentials)
	}

	dRootAt0 := []byte("439d4fe8d685a0a752a95b390c63a612fd29056dbd648afee5ffbc23c3f8a62e")
	dRootAt0B := make([]byte, hex.DecodedLen(len(dRootAt0)))
	_, err = hex.Decode(dRootAt0B, dRootAt0)
	require.NoError(t, err)
	if !bytes.Equal(depositDataRoots[0][:], dRootAt0B) {
		t.Fatalf("incorrect deposit data root, wanted %x but received %x", dRootAt0B, depositDataRoots[0])
	}

	sigAt0 := []byte("ac5f85c7041af233ce20cff35e2da4a9529314a56a788e9af74769dc7143d98453cc989fca95c4191a75b79e4d3bd55b0419f828e88bb9754b3aefe8c6d435df375e2f0ce8a645328c8b6664d680ce3359e9dd0307d4393236c4c5eda7390b95")
	sigAt0B := make([]byte, hex.DecodedLen(len(sigAt0)))
	_, err = hex.Decode(sigAt0B, sigAt0)
	require.NoError(t, err)
	if !bytes.Equal(deposits[0].Data.Signature, sigAt0B) {
		t.Fatalf("incorrect signature, wanted %x but received %x", sigAt0B, deposits[0].Data.Signature)
	}

	entries = 1024
	resetCache()
	deposits, privKeys, err = DeterministicDepositsAndKeys(uint64(entries))
	require.NoError(t, err)
	_, depositDataRoots, err = DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	if len(deposits) != entries {
		t.Fatalf("incorrect number of deposits returned, wanted %d but received %d", entries, len(deposits))
	}
	if len(privKeys) != entries {
		t.Fatalf("incorrect number of private keys returned, wanted %d but received %d", entries, len(privKeys))
	}
	// Ensure 0  has not changed
	if !bytes.Equal(deposits[0].Data.PublicKey, expectedPublicKeyAt0) {
		t.Fatalf("incorrect public key, wanted %x but received %x", expectedPublicKeyAt0, deposits[0].Data.PublicKey)
	}
	if !bytes.Equal(deposits[0].Data.WithdrawalCredentials, expectedWithdrawalCredentialsAt0) {
		t.Fatalf("incorrect withdrawal credentials, wanted %x but received %x", expectedWithdrawalCredentialsAt0, deposits[0].Data.WithdrawalCredentials)
	}
	if !bytes.Equal(depositDataRoots[0][:], dRootAt0B) {
		t.Fatalf("incorrect deposit data root, wanted %x but received %x", dRootAt0B, depositDataRoots[0])
	}
	if !bytes.Equal(deposits[0].Data.Signature, sigAt0B) {
		t.Fatalf("incorrect signature, wanted %x but received %x", sigAt0B, deposits[0].Data.Signature)
	}
	expectedPublicKeyAt1023 := []byte{0x81, 0x2b, 0x93, 0x5e, 0xc8, 0x4b, 0x0e, 0x9a, 0x83, 0x95, 0x55, 0xaf, 0x33, 0x60, 0xca, 0xfb, 0x83, 0x1b, 0xd6, 0x12, 0xcf, 0xa2, 0x2e, 0x25, 0xea, 0xb0, 0x3c, 0xf5, 0xfd, 0xb0, 0x2a, 0xf5, 0x2b, 0xa4, 0x01, 0x7a, 0xee, 0xa8, 0x8a, 0x2f, 0x62, 0x2c, 0x78, 0x6e, 0x7f, 0x47, 0x6f, 0x4b}
	if !bytes.Equal(deposits[1023].Data.PublicKey, expectedPublicKeyAt1023) {
		t.Fatalf("incorrect public key, wanted %x but received %x", expectedPublicKeyAt1023, deposits[1023].Data.PublicKey)
	}
	expectedWithdrawalCredentialsAt1023 := []byte{0xCB, 0xA7, 0x7E, 0xAC, 0xE1, 0x0, 0xAA, 0xB6, 0x67, 0x86, 0x12, 0xEA, 0xC4, 0xE3, 0x57, 0x93, 0x2D, 0x70, 0xD4, 0x96}
	if !bytes.Equal(deposits[1023].Data.WithdrawalCredentials, expectedWithdrawalCredentialsAt1023) {
		t.Fatalf("incorrect withdrawal credentials, wanted %x but received %x", expectedWithdrawalCredentialsAt1023, deposits[1023].Data.WithdrawalCredentials)
	}
	dRootAt1023 := []byte("f958fb588cda5d320b3b0cbed8ee6af7b517d8b4390cb9a7e7840c34e8c9135b")
	dRootAt1023B := make([]byte, hex.DecodedLen(len(dRootAt1023)))
	_, err = hex.Decode(dRootAt1023B, dRootAt1023)
	require.NoError(t, err)
	if !bytes.Equal(depositDataRoots[1023][:], dRootAt1023B) {
		t.Fatalf("incorrect deposit data root, wanted %x but received %x", dRootAt1023B, depositDataRoots[1023])
	}
	sigAt1023 := []byte("80c8993e5e40f2e0abe7a32f3051654a410b6b1541eb27cd4e4f48b1a67d0c577fb8f80f2cf6e348ec448fcff0a4256b197498fefc308869b4f7434b1497b9c4416fc9e280ab66248e886992794c63bbbaf776821fa777bd7f5a367dee128b7e")
	sigAt1023B := make([]byte, hex.DecodedLen(len(sigAt1023)))
	_, err = hex.Decode(sigAt1023B, sigAt1023)
	require.NoError(t, err)
	if !bytes.Equal(deposits[1023].Data.Signature, sigAt1023B) {
		t.Fatalf("incorrect signature, wanted %x but received %x", sigAt1023B, deposits[1023].Data.Signature)
	}
}

func TestDepositsWithBalance_MatchesDeterministic(t *testing.T) {
	entries := 64
	resetCache()
	balances := make([]uint64, entries)
	for i := 0; i < entries; i++ {
		balances[i] = params.BeaconConfig().MaxEffectiveBalance
	}
	deposits, depositTrie, err := DepositsWithBalance(balances)
	require.NoError(t, err)
	_, depositDataRoots, err := DepositTrieSubset(depositTrie, entries)
	require.NoError(t, err)

	determDeposits, _, err := DeterministicDepositsAndKeys(uint64(entries))
	require.NoError(t, err)
	_, determDepositDataRoots, err := DeterministicDepositTrie(entries)
	require.NoError(t, err)

	for i := 0; i < entries; i++ {
		if !proto.Equal(deposits[i], determDeposits[i]) {
			t.Errorf("Expected deposit %d to match", i)
		}
		if !bytes.Equal(depositDataRoots[i][:], determDepositDataRoots[i][:]) {
			t.Errorf("Expected deposit root %d to match", i)
		}
	}
}

func TestDepositsWithBalance_MatchesDeterministic_Cached(t *testing.T) {
	entries := 32
	resetCache()
	// Cache half of the deposit cache.
	_, _, err := DeterministicDepositsAndKeys(uint64(entries))
	require.NoError(t, err)
	_, _, err = DeterministicDepositTrie(entries)
	require.NoError(t, err)

	// Generate balanced deposits with half cache.
	entries = 64
	balances := make([]uint64, entries)
	for i := 0; i < entries; i++ {
		balances[i] = params.BeaconConfig().MaxEffectiveBalance
	}
	deposits, depositTrie, err := DepositsWithBalance(balances)
	require.NoError(t, err)
	_, depositDataRoots, err := DepositTrieSubset(depositTrie, entries)
	require.NoError(t, err)

	// Get 64 standard deposits.
	determDeposits, _, err := DeterministicDepositsAndKeys(uint64(entries))
	require.NoError(t, err)
	_, determDepositDataRoots, err := DeterministicDepositTrie(entries)
	require.NoError(t, err)

	for i := 0; i < entries; i++ {
		if !proto.Equal(deposits[i], determDeposits[i]) {
			t.Errorf("Expected deposit %d to match", i)
		}
		if !bytes.Equal(depositDataRoots[i][:], determDepositDataRoots[i][:]) {
			t.Errorf("Expected deposit root %d to match", i)
		}
	}
}

func TestSetupInitialDeposits_1024Entries_PartialDeposits(t *testing.T) {
	entries := 1
	resetCache()
	balances := make([]uint64, entries)
	for i := 0; i < entries; i++ {
		balances[i] = params.BeaconConfig().MaxEffectiveBalance / 2
	}
	deposits, depositTrie, err := DepositsWithBalance(balances)
	require.NoError(t, err)
	_, depositDataRoots, err := DepositTrieSubset(depositTrie, entries)
	require.NoError(t, err)

	if len(deposits) != entries {
		t.Fatalf("incorrect number of deposits returned, wanted %d but received %d", entries, len(deposits))
	}
	expectedPublicKeyAt0 := []byte{0xa9, 0x9a, 0x76, 0xed, 0x77, 0x96, 0xf7, 0xbe, 0x22, 0xd5, 0xb7, 0xe8, 0x5d, 0xee, 0xb7, 0xc5, 0x67, 0x7e, 0x88, 0xe5, 0x11, 0xe0, 0xb3, 0x37, 0x61, 0x8f, 0x8c, 0x4e, 0xb6, 0x13, 0x49, 0xb4, 0xbf, 0x2d, 0x15, 0x3f, 0x64, 0x9f, 0x7b, 0x53, 0x35, 0x9f, 0xe8, 0xb9, 0x4a, 0x38, 0xe4, 0x4c}
	if !bytes.Equal(deposits[0].Data.PublicKey, expectedPublicKeyAt0) {
		t.Fatalf("incorrect public key, wanted %x but received %x", expectedPublicKeyAt0, deposits[0].Data.PublicKey)
	}
	expectedWithdrawalCredentialsAt0 := []byte{0x54, 0x7d, 0xcb, 0xa5, 0xba, 0xc1, 0x6a, 0x89, 0x10, 0x8b, 0x6b, 0x6a, 0x1f, 0xe3, 0x69, 0x5d, 0x1a, 0x87, 0x4a, 0x0b}
	if !bytes.Equal(deposits[0].Data.WithdrawalCredentials, expectedWithdrawalCredentialsAt0) {
		t.Fatalf("incorrect withdrawal credentials, wanted %x but received %x", expectedWithdrawalCredentialsAt0, deposits[0].Data.WithdrawalCredentials)
	}
	dRootAt0 := []byte("55834cd732fa7d24e4333be32ff0e8e7152e52c19035237e83fe4f4088124119")
	dRootAt0B := make([]byte, hex.DecodedLen(len(dRootAt0)))
	_, err = hex.Decode(dRootAt0B, dRootAt0)
	require.NoError(t, err)
	if !bytes.Equal(depositDataRoots[0][:], dRootAt0B) {
		t.Fatalf("incorrect deposit data root, wanted %#x but received %#x", dRootAt0B, depositDataRoots[0])
	}

	sigAt0 := []byte("ac5f85c7041af233ce20cff35e2da4a9529314a56a788e9af74769dc7143d98453cc989fca95c4191a75b79e4d3bd55b0419f828e88bb9754b3aefe8c6d435df375e2f0ce8a645328c8b6664d680ce3359e9dd0307d4393236c4c5eda7390b95")
	sigAt0B := make([]byte, hex.DecodedLen(len(sigAt0)))
	_, err = hex.Decode(sigAt0B, sigAt0)
	require.NoError(t, err)
	if !bytes.Equal(deposits[0].Data.Signature, sigAt0B) {
		t.Fatalf("incorrect signature, wanted %#x but received %#x", sigAt0B, deposits[0].Data.Signature)
	}

	entries = 1024
	resetCache()
	balances = make([]uint64, entries)
	for i := 0; i < entries; i++ {
		balances[i] = params.BeaconConfig().MaxEffectiveBalance / 2
	}
	deposits, depositTrie, err = DepositsWithBalance(balances)
	require.NoError(t, err)
	_, depositDataRoots, err = DepositTrieSubset(depositTrie, entries)
	require.NoError(t, err)
	if len(deposits) != entries {
		t.Fatalf("incorrect number of deposits returned, wanted %d but received %d", entries, len(deposits))
	}
	// Ensure 0  has not changed
	if !bytes.Equal(deposits[0].Data.PublicKey, expectedPublicKeyAt0) {
		t.Fatalf("incorrect public key, wanted %x but received %x", expectedPublicKeyAt0, deposits[0].Data.PublicKey)
	}
	if !bytes.Equal(deposits[0].Data.WithdrawalCredentials, expectedWithdrawalCredentialsAt0) {
		t.Fatalf("incorrect withdrawal credentials, wanted %x but received %x", expectedWithdrawalCredentialsAt0, deposits[0].Data.WithdrawalCredentials)
	}
	if !bytes.Equal(depositDataRoots[0][:], dRootAt0B) {
		t.Fatalf("incorrect deposit data root, wanted %x but received %x", dRootAt0B, depositDataRoots[0])
	}
	if !bytes.Equal(deposits[0].Data.Signature, sigAt0B) {
		t.Fatalf("incorrect signature, wanted %x but received %x", sigAt0B, deposits[0].Data.Signature)
	}
	expectedPublicKeyAt1023 := []byte{0x81, 0x2b, 0x93, 0x5e, 0xc8, 0x4b, 0x0e, 0x9a, 0x83, 0x95, 0x55, 0xaf, 0x33, 0x60, 0xca, 0xfb, 0x83, 0x1b, 0xd6, 0x12, 0xcf, 0xa2, 0x2e, 0x25, 0xea, 0xb0, 0x3c, 0xf5, 0xfd, 0xb0, 0x2a, 0xf5, 0x2b, 0xa4, 0x01, 0x7a, 0xee, 0xa8, 0x8a, 0x2f, 0x62, 0x2c, 0x78, 0x6e, 0x7f, 0x47, 0x6f, 0x4b}
	if !bytes.Equal(deposits[1023].Data.PublicKey, expectedPublicKeyAt1023) {
		t.Fatalf("incorrect public key, wanted %x but received %x", expectedPublicKeyAt1023, deposits[1023].Data.PublicKey)
	}
	expectedWithdrawalCredentialsAt1023 := []byte{0xcb, 0xa7, 0x7e, 0xac, 0xe1, 0x00, 0xaa, 0xb6, 0x67, 0x86, 0x12, 0xea, 0xc4, 0xe3, 0x57, 0x93, 0x2d, 0x70, 0xd4, 0x96}
	if !bytes.Equal(deposits[1023].Data.WithdrawalCredentials, expectedWithdrawalCredentialsAt1023) {
		t.Fatalf("incorrect withdrawal credentials, wanted %x but received %x", expectedWithdrawalCredentialsAt1023, deposits[1023].Data.WithdrawalCredentials)
	}
	dRootAt1023 := []byte("19207e22bd575e9feb040253c1c431d8212a117184cff1a0809b2454fb509ef6")
	dRootAt1023B := make([]byte, hex.DecodedLen(len(dRootAt1023)))
	_, err = hex.Decode(dRootAt1023B, dRootAt1023)
	require.NoError(t, err)
	if !bytes.Equal(depositDataRoots[1023][:], dRootAt1023B) {
		t.Fatalf("incorrect deposit data root, wanted %#x but received %#x", dRootAt1023B, depositDataRoots[1023])
	}
	sigAt1023 := []byte("80c8993e5e40f2e0abe7a32f3051654a410b6b1541eb27cd4e4f48b1a67d0c577fb8f80f2cf6e348ec448fcff0a4256b197498fefc308869b4f7434b1497b9c4416fc9e280ab66248e886992794c63bbbaf776821fa777bd7f5a367dee128b7e")
	sigAt1023B := make([]byte, hex.DecodedLen(len(sigAt1023)))
	_, err = hex.Decode(sigAt1023B, sigAt1023)
	require.NoError(t, err)
	if !bytes.Equal(deposits[1023].Data.Signature, sigAt1023B) {
		t.Fatalf("incorrect signature, wanted %#x but received %#x", sigAt1023B, deposits[1023].Data.Signature)
	}
}

func TestDeterministicGenesisState_100Validators(t *testing.T) {
	validatorCount := uint64(100)
	beaconState, privKeys := DeterministicGenesisState(t, validatorCount)
	activeValidators, err := helpers.ActiveValidatorCount(context.Background(), beaconState, 0)
	require.NoError(t, err)

	// lint:ignore uintcast -- test code
	if len(privKeys) != int(validatorCount) {
		t.Fatalf("expected amount of private keys %d to match requested amount of validators %d", len(privKeys), validatorCount)
	}
	if activeValidators != validatorCount {
		t.Fatalf("expected validators in state %d to match requested amount %d", activeValidators, validatorCount)
	}
}

func TestDepositTrieFromDeposits(t *testing.T) {
	deposits, _, err := DeterministicDepositsAndKeys(100)
	require.NoError(t, err)
	eth1Data, err := DeterministicEth1Data(len(deposits))
	require.NoError(t, err)

	depositTrie, _, err := DepositTrieFromDeposits(deposits)
	require.NoError(t, err)

	root := depositTrie.HashTreeRoot()
	if !bytes.Equal(root[:], eth1Data.DepositRoot) {
		t.Fatal("expected deposit trie root to equal eth1data deposit root")
	}
}
