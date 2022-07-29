package main

import (
	"fmt"
	"sync"

	"github.com/prysmaticlabs/prysm/encoding/bytesutil"
	v1alpha1 "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1/block"
	"github.com/waterfall-foundation/gwat/crypto"
)

type votes struct {
	l sync.RWMutex

	candidates map[[32]byte]uint
	hashes     map[[32]byte]uint
	roots      map[[32]byte]uint
	counts     map[uint64]uint
	votes      map[[32]byte]*v1alpha1.Eth1Data
	voteCounts map[[32]byte]uint
	total      uint
}

func NewVotes() *votes {
	return &votes{
		candidates: make(map[[32]byte]uint),
		hashes:     make(map[[32]byte]uint),
		roots:      make(map[[32]byte]uint),
		counts:     make(map[uint64]uint),
		votes:      make(map[[32]byte]*v1alpha1.Eth1Data),
		voteCounts: make(map[[32]byte]uint),
	}
}

func (v *votes) Insert(blk block.BeaconBlock) {
	v.l.Lock()
	defer v.l.Unlock()

	e1d := blk.Body().Eth1Data()
	htr, err := e1d.HashTreeRoot()
	if err != nil {
		panic(err)
	}
	finId := crypto.Keccak256Hash(e1d.Candidates)
	v.candidates[bytesutil.ToBytes32(finId.Bytes())]++
	v.hashes[bytesutil.ToBytes32(e1d.BlockHash)]++
	v.roots[bytesutil.ToBytes32(e1d.DepositRoot)]++
	v.counts[e1d.DepositCount]++
	v.votes[htr] = e1d
	v.voteCounts[htr]++
	v.total++
}

func (v *votes) Report() string {
	v.l.RLock()
	defer v.l.RUnlock()
	format := `====Eth1Data Voting Report====

Total votes: %d

Hashed Candidates
%s
Block Hashes
%s
Deposit Roots
%s
Deposit Counts
%s
Votes
%s
`
	var hashedCandidates string
	for r, cnt := range v.candidates {
		hashedCandidates += fmt.Sprintf("%#x=%d\n", r, cnt)
	}
	var blockHashes string
	for r, cnt := range v.hashes {
		blockHashes += fmt.Sprintf("%#x=%d\n", r, cnt)
	}
	var depositRoots string
	for r, cnt := range v.roots {
		depositRoots += fmt.Sprintf("%#x=%d\n", r, cnt)
	}
	var depositCounts string
	for dc, cnt := range v.counts {
		depositCounts += fmt.Sprintf("%d=%d\n", dc, cnt)
	}
	var votes string
	for htr, e1d := range v.votes {
		votes += fmt.Sprintf("%s=%d\n", e1d.String(), v.voteCounts[htr])
	}

	return fmt.Sprintf(
		format,
		v.total,
		hashedCandidates,
		blockHashes,
		depositRoots,
		depositCounts,
		votes,
	)
}
