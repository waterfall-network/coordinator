package protoarray

import (
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/require"
)

func TestStore_GetFork(t *testing.T) {
	f := &ForkChoice{store: &Store{}}
	f.store.canonicalNodes = map[[32]byte]bool{}
	f.store.nodesIndices = map[[32]byte]uint64{
		[32]byte{'a'}: 0,
		[32]byte{'b'}: 1,
		[32]byte{'c'}: 2,
		[32]byte{'d'}: 3,
		[32]byte{'e'}: 4,
		[32]byte{'f'}: 5,
		[32]byte{'g'}: 6,
		[32]byte{'h'}: 7,
		[32]byte{'i'}: 8,
		[32]byte{'j'}: 9,
		[32]byte{'k'}: 10,
	}
	f.store.nodes = []*Node{
		//fork 0
		{slot: 1, root: [32]byte{'a'}, parent: NonExistentNode},
		{slot: 2, root: [32]byte{'b'}, parent: 0},
		{slot: 3, root: [32]byte{'c'}, parent: 1},
		{slot: 4, root: [32]byte{'d'}, parent: 2},
		{slot: 5, root: [32]byte{'e'}, parent: 3},
		{slot: 6, root: [32]byte{'f'}, parent: 4},
		//fork 1
		{slot: 7, root: [32]byte{'g'}, parent: 2},
		{slot: 8, root: [32]byte{'h'}, parent: 6},
		//fork 2
		{slot: 9, root: [32]byte{'i'}, parent: 3},
		{slot: 10, root: [32]byte{'j'}, parent: 8},
		{slot: 11, root: [32]byte{'k'}, parent: 9},
	}
	want := &Fork{
		roots: [][32]byte{
			[32]byte{'h'},
			[32]byte{'g'},
			[32]byte{'c'},
			[32]byte{'b'},
			[32]byte{'a'},
		},
		nodesMap: map[[32]byte]*Node{
			[32]byte{'h'}: {slot: 8, root: [32]byte{'h'}, parent: 6},
			[32]byte{'g'}: {slot: 7, root: [32]byte{'g'}, parent: 2},
			[32]byte{'c'}: {slot: 3, root: [32]byte{'c'}, parent: 1},
			[32]byte{'b'}: {slot: 2, root: [32]byte{'b'}, parent: 0},
			[32]byte{'a'}: {slot: 1, root: [32]byte{'a'}, parent: NonExistentNode},
		},
	}
	got := f.GetFork([32]byte{'h'})
	require.DeepEqual(t, want, got)
}

func TestStore_GetForks(t *testing.T) {
	f := &ForkChoice{store: &Store{}}
	f.store.canonicalNodes = map[[32]byte]bool{}
	f.store.nodesIndices = map[[32]byte]uint64{
		[32]byte{'a'}: 0,
		[32]byte{'b'}: 1,
		[32]byte{'c'}: 2,
		[32]byte{'d'}: 3,
		[32]byte{'e'}: 4,
		[32]byte{'f'}: 5,
		[32]byte{'g'}: 6,
		[32]byte{'h'}: 7,
		[32]byte{'i'}: 8,
		[32]byte{'j'}: 9,
		[32]byte{'k'}: 10,
	}
	f.store.nodes = []*Node{
		//fork 0
		{slot: 1, root: [32]byte{'a'}, parent: NonExistentNode},
		{slot: 2, root: [32]byte{'b'}, parent: 0},
		{slot: 3, root: [32]byte{'c'}, parent: 1},
		{slot: 4, root: [32]byte{'d'}, parent: 2},
		{slot: 5, root: [32]byte{'e'}, parent: 3},
		{slot: 6, root: [32]byte{'f'}, parent: 4},
		//fork 1
		{slot: 7, root: [32]byte{'g'}, parent: 2},
		{slot: 8, root: [32]byte{'h'}, parent: 6},
		//fork 2
		{slot: 9, root: [32]byte{'i'}, parent: 3},
		{slot: 10, root: [32]byte{'j'}, parent: 8},
		{slot: 11, root: [32]byte{'k'}, parent: 9},
	}
	want := []*Fork{
		&Fork{
			roots: [][32]byte{
				[32]byte{'k'},
				[32]byte{'j'},
				[32]byte{'i'},
				[32]byte{'d'},
				[32]byte{'c'},
				[32]byte{'b'},
				[32]byte{'a'},
			},
			nodesMap: map[[32]byte]*Node{
				[32]byte{'k'}: {slot: 11, root: [32]byte{'k'}, parent: 9},
				[32]byte{'j'}: {slot: 10, root: [32]byte{'j'}, parent: 8},
				[32]byte{'i'}: {slot: 9, root: [32]byte{'i'}, parent: 3},
				[32]byte{'d'}: {slot: 4, root: [32]byte{'d'}, parent: 2},
				[32]byte{'c'}: {slot: 3, root: [32]byte{'c'}, parent: 1},
				[32]byte{'b'}: {slot: 2, root: [32]byte{'b'}, parent: 0},
				[32]byte{'a'}: {slot: 1, root: [32]byte{'a'}, parent: NonExistentNode},
			},
		},
		&Fork{
			roots: [][32]byte{
				[32]byte{'h'},
				[32]byte{'g'},
				[32]byte{'c'},
				[32]byte{'b'},
				[32]byte{'a'},
			},
			nodesMap: map[[32]byte]*Node{
				[32]byte{'h'}: {slot: 8, root: [32]byte{'h'}, parent: 6},
				[32]byte{'g'}: {slot: 7, root: [32]byte{'g'}, parent: 2},
				[32]byte{'c'}: {slot: 3, root: [32]byte{'c'}, parent: 1},
				[32]byte{'b'}: {slot: 2, root: [32]byte{'b'}, parent: 0},
				[32]byte{'a'}: {slot: 1, root: [32]byte{'a'}, parent: NonExistentNode},
			},
		},
		&Fork{
			roots: [][32]byte{
				[32]byte{'f'},
				[32]byte{'e'},
				[32]byte{'d'},
				[32]byte{'c'},
				[32]byte{'b'},
				[32]byte{'a'},
			},
			nodesMap: map[[32]byte]*Node{
				[32]byte{'f'}: {slot: 6, root: [32]byte{'f'}, parent: 4},
				[32]byte{'e'}: {slot: 5, root: [32]byte{'e'}, parent: 3},
				[32]byte{'d'}: {slot: 4, root: [32]byte{'d'}, parent: 2},
				[32]byte{'c'}: {slot: 3, root: [32]byte{'c'}, parent: 1},
				[32]byte{'b'}: {slot: 2, root: [32]byte{'b'}, parent: 0},
				[32]byte{'a'}: {slot: 1, root: [32]byte{'a'}, parent: NonExistentNode},
			},
		},
	}

	got := f.GetForks()
	require.DeepEqual(t, want, got)
}

func TestStore_GetFirstFork(t *testing.T) {
	f := &ForkChoice{store: &Store{}}
	f.store.canonicalNodes = map[[32]byte]bool{}
	f.store.nodesIndices = map[[32]byte]uint64{
		[32]byte{'a'}: 0,
		[32]byte{'b'}: 1,
		[32]byte{'c'}: 2,
		[32]byte{'d'}: 3,
		[32]byte{'e'}: 4,
		[32]byte{'f'}: 5,
		[32]byte{'g'}: 6,
		[32]byte{'h'}: 7,
		[32]byte{'i'}: 8,
		[32]byte{'j'}: 9,
		[32]byte{'k'}: 10,
	}
	f.store.nodes = []*Node{
		//fork 0
		{slot: 1, root: [32]byte{'a'}, parent: NonExistentNode},
		{slot: 2, root: [32]byte{'b'}, parent: 0},
		{slot: 3, root: [32]byte{'c'}, parent: 1},
		{slot: 4, root: [32]byte{'d'}, parent: 2},
		{slot: 5, root: [32]byte{'e'}, parent: 3},
		{slot: 6, root: [32]byte{'f'}, parent: 4},
		//fork 1
		{slot: 7, root: [32]byte{'g'}, parent: 2},
		{slot: 8, root: [32]byte{'h'}, parent: 6},
		//fork 2
		{slot: 9, root: [32]byte{'i'}, parent: 3},
		{slot: 10, root: [32]byte{'j'}, parent: 8},
		{slot: 11, root: [32]byte{'k'}, parent: 9},
	}
	want := &Node{slot: 3, root: [32]byte{'c'}, parent: 1}
	got := f.GetCommonAncestor()
	require.DeepEqual(t, want, got)
}
