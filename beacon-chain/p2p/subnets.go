package p2p

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/beacon-chain/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	mathutil "gitlab.waterfall.network/waterfall/protocol/coordinator/math"
	pb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/wrapper"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/enode"
	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/enr"
	"go.opencensus.io/trace"
)

var attestationSubnetCount = params.BeaconNetworkConfig().AttestationSubnetCount
var syncCommsSubnetCount = params.BeaconConfig().SyncCommitteeSubnetCount

var attSubnetEnrKey = params.BeaconNetworkConfig().AttSubnetKey
var syncCommsSubnetEnrKey = params.BeaconNetworkConfig().SyncCommsSubnetKey

// The value used with the subnet, inorder
// to create an appropriate key to retrieve
// the relevant lock. This is used to differentiate
// sync subnets from attestation subnets. This is deliberately
// chosen as more than 64(attestation subnet count).
const syncLockerVal = 100

// FindPeersWithSubnet performs a network search for peers
// subscribed to a particular subnet. Then we try to connect
// with those peers. This method will block until the required amount of
// peers are found, the method only exits in the event of context timeouts.
func (s *Service) FindPeersWithSubnet(ctx context.Context, topic string,
	index uint64, threshold int) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "p2p.FindPeersWithSubnet")
	defer span.End()

	span.AddAttributes(trace.Int64Attribute("index", int64(index))) // lint:ignore uintcast -- It's safe to do this for tracing.

	//todo RM
	log.WithFields(logrus.Fields{
		"curSlot":              slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"idx":                  index,
		"digest":               fmt.Sprintf("%s", topic),
		"threshold":            threshold,
		"s.dv5Listener == nil": s.dv5Listener == nil,
	}).Info("Prevote: FindPeersWithSubnet: 000")

	if s.dv5Listener == nil {
		// return if discovery isn't set
		return false, nil
	}

	topic += s.Encoding().ProtocolSuffix()
	iterator := s.dv5Listener.RandomNodes()
	switch {
	case strings.Contains(topic, GossipAttestationMessage):
		iterator = filterNodes(ctx, iterator, s.filterPeerForAttSubnet(index))
	case strings.Contains(topic, GossipSyncCommitteeMessage):
		iterator = filterNodes(ctx, iterator, s.filterPeerForSyncSubnet(index))
	default:
		return false, errors.New("no subnet exists for provided topic")
	}
	currNum := len(s.pubsub.ListPeers(topic))

	//todo RM
	log.WithFields(logrus.Fields{
		"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"idx":       index,
		"digest":    fmt.Sprintf("%s", topic),
		"threshold": threshold,
		"currNum":   currNum,
	}).Info("Prevote: FindPeersWithSubnet: 111")

	wg := new(sync.WaitGroup)
	actionCounter := 0

	//TODO ++++++++
	//for {
	//	if err := ctx.Err(); err != nil {
	//
	//		//todo RM
	//		log.WithFields(logrus.Fields{
	//			"curSlot": slots.CurrentSlot(uint64(s.genesisTime.Unix())),
	//			"idx":     index,
	//			"currNum": currNum,
	//		}).WithError(err).Error("Prevote: FindPeersWithSubnet: CYCLi-000 err")
	//
	//		return false, errors.Errorf("unable to find requisite number of peers for topic %s - "+
	//			"only %d out of %d peers were able to be found", topic, currNum, threshold)
	//	}
	//	if currNum >= threshold {
	//
	//		//todo RM
	//		log.WithFields(logrus.Fields{
	//			"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
	//			"idx":       index,
	//			"threshold": threshold,
	//			"currNum":   currNum,
	//		}).Info("Prevote: FindPeersWithSubnet: CYCLi-999 SUCCESS")
	//
	//		break
	//	}
	//TODO ++++++++

	//todo RM
	log.WithFields(logrus.Fields{
		"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"idx":       index,
		"threshold": threshold,
		"currNum":   currNum,
	}).Info("Prevote: FindPeersWithSubnet: CYCLi-00000")

	//todo recomment rm
	//nodes := enode.ReadNodes(iterator, int(params.BeaconNetworkConfig().MinimumPeersInSubnetSearch))
	nodes := ReadNodes(iterator, int(params.BeaconNetworkConfig().MinimumPeersInSubnetSearch), index)

	//todo RM
	log.WithFields(logrus.Fields{
		"len(nodes)": len(nodes),
		"curSlot":    slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"idx":        index,
		"threshold":  threshold,
		"currNum":    currNum,
	}).Info("Prevote: FindPeersWithSubnet: CYCLi-111")

	// if no nodes
	if len(nodes) == 0 {
		time.Sleep(time.Duration(20) * time.Millisecond)
	}

	for j, node := range nodes {

		//todo RM
		log.WithFields(logrus.Fields{
			"j":         j,
			"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
			"idx":       index,
			"threshold": threshold,
			"currNum":   currNum,
		}).Info("Prevote: FindPeersWithSubnet: CYCLij-000")

		info, _, err := convertToAddrInfo(node)
		if err != nil {

			//todo RM
			log.WithFields(logrus.Fields{
				"j":         j,
				"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
				"idx":       index,
				"threshold": threshold,
				"currNum":   currNum,
			}).WithError(err).Error("Prevote: FindPeersWithSubnet: CYCLij-111 CONTINUE")

			continue
		}
		wg.Add(1)
		actionCounter++
		go func() {

			//todo RM
			log.WithFields(logrus.Fields{
				"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
				"idx":       index,
				"threshold": threshold,
				"currNum":   currNum,
				"nodeInf":   fmt.Sprintf("%s", info.String()),
			}).Info("Prevote: FindPeersWithSubnet: CYCLij-222")

			if err := s.connectWithPeer(ctx, *info); err != nil {

				//todo RM
				log.WithFields(logrus.Fields{
					"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
					"idx":       index,
					"threshold": threshold,
					"currNum":   currNum,
					"nodeInf":   fmt.Sprintf("%s", info.String()),
				}).WithError(err).Error("Prevote: FindPeersWithSubnet: CYCLij-333 ERR")

				log.WithError(err).Errorf("Could not connect with peer %s", info.String())
			}

			//todo RM
			log.WithFields(logrus.Fields{
				"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
				"idx":       index,
				"threshold": threshold,
				"currNum":   currNum,
				"nodeInf":   fmt.Sprintf("%s", info.String()),
			}).Info("Prevote: FindPeersWithSubnet: CYCLij-444")

			wg.Done()
			actionCounter--
		}()
	}

	//todo RM
	log.WithFields(logrus.Fields{
		"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"idx":       index,
		"threshold": threshold,
		"currNum":   currNum,
	}).Info("Prevote: FindPeersWithSubnet: CYCLij-222")

	// Wait for all dials to be completed.
	wg.Wait()

	//todo RM
	log.WithFields(logrus.Fields{
		"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"idx":       index,
		"threshold": threshold,
		"currNum":   currNum,
	}).Info("Prevote: FindPeersWithSubnet: CYCLij-333")

	currNum = len(s.pubsub.ListPeers(topic))

	//todo RM
	log.WithFields(logrus.Fields{
		"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
		"idx":       index,
		"threshold": threshold,
		"currNum":   currNum,
	}).Info("Prevote: FindPeersWithSubnet: CYCLij-333")

	//TODO ++++++++
	//}

	return true, nil
}

// ReadNodes reads at most n nodes from the given iterator. The return value contains no
// duplicates and no nil values. To prevent looping indefinitely for small repeating node
// sequences, this function calls Next at most n times.
func ReadNodes(it enode.Iterator, n int, index uint64) []*enode.Node {
	seen := make(map[enode.ID]*enode.Node, n)

	//todo RM
	log.WithFields(logrus.Fields{
		"idx": index,
		"n":   n,
	}).Info("Prevote: ReadNodes: 000")

	for i := 0; i < n && it.Next(); i++ {

		//todo RM
		log.WithFields(logrus.Fields{
			"i":   i,
			"idx": index,
			"n":   n,
		}).Info("Prevote: ReadNodes: 111")

		// Remove duplicates, keeping the node with higher seq.
		node := it.Node()
		prevNode, ok := seen[node.ID()]

		//todo RM
		log.WithFields(logrus.Fields{
			"i":            i,
			"idx":          index,
			"n":            n,
			"continue":     ok && prevNode.Seq() > node.Seq(),
			"ok":           ok,
			"prevNode.Seq": prevNode.Seq(),
			"node.Seq":     node.Seq(),
		}).Info("Prevote: ReadNodes: 222")

		if ok && prevNode.Seq() > node.Seq() {
			continue
		}
		seen[node.ID()] = node

		//todo RM
		log.WithFields(logrus.Fields{
			"i":   i,
			"idx": index,
			"n":   n,
		}).Info("Prevote: ReadNodes: 333")
	}

	log.WithFields(logrus.Fields{
		"idx":       index,
		"n":         n,
		"len(seen)": len(seen),
	}).Info("Prevote: ReadNodes: 555")

	result := make([]*enode.Node, 0, len(seen))
	for i, node := range seen {

		log.WithFields(logrus.Fields{
			"i":    i,
			"idx":  index,
			"n":    n,
			"node": node.ID(),
		}).Info("Prevote: ReadNodes: 666 iter")

		result = append(result, node)
	}

	log.WithFields(logrus.Fields{
		"idx":         index,
		"n":           n,
		"len(result)": len(result),
	}).Info("Prevote: ReadNodes: 777")

	return result
}

// returns a method with filters peers specifically for a particular attestation subnet.
func (s *Service) filterPeerForAttSubnet(index uint64) func(node *enode.Node) bool {
	return func(node *enode.Node) bool {

		//todo RM
		log.WithFields(logrus.Fields{
			"curSlot": slots.CurrentSlot(uint64(s.genesisTime.Unix())),
			"idx":     index,
			"node":    node.ID(),
		}).Info("Prevote: filterPeerForAttSubnet: 000")

		if !s.filterPeer(node) {

			//todo RM
			log.WithFields(logrus.Fields{
				"curSlot": slots.CurrentSlot(uint64(s.genesisTime.Unix())),
				"idx":     index,
				"node":    node.ID(),
			}).Info("Prevote: filterPeerForAttSubnet: 111 ret F")

			//if index > 10 {
			//	panic("Prevote: filterPeerForAttSubnet: 111 ret F")
			//}

			return false
		}

		//todo RM
		log.WithFields(logrus.Fields{
			"curSlot": slots.CurrentSlot(uint64(s.genesisTime.Unix())),
			"idx":     index,
			"node":    node.ID(),
		}).Info("Prevote: filterPeerForAttSubnet: 222")

		subnets, err := attSubnets(node.Record())
		if err != nil {

			//todo RM
			log.WithFields(logrus.Fields{
				"curSlot": slots.CurrentSlot(uint64(s.genesisTime.Unix())),
				"idx":     index,
				"node":    node.ID(),
			}).Info("Prevote: filterPeerForAttSubnet: 333 ret F")

			return false
		}

		//todo RM
		log.WithFields(logrus.Fields{
			"curSlot": slots.CurrentSlot(uint64(s.genesisTime.Unix())),
			"idx":     index,
			"node":    node.ID(),
		}).Info("Prevote: filterPeerForAttSubnet: 444")

		indExists := false
		for i, comIdx := range subnets {

			//todo RM
			log.WithFields(logrus.Fields{
				"i":             i,
				"comIdx":        comIdx,
				"curSlot":       slots.CurrentSlot(uint64(s.genesisTime.Unix())),
				"idx":           index,
				"node":          node.ID(),
				"comIdx==index": comIdx == index,
			}).Info("Prevote: filterPeerForAttSubnet: 444")

			if comIdx == index {
				indExists = true
				break
			}
		}

		//todo RM
		log.WithFields(logrus.Fields{
			"indExists": indExists,
			"curSlot":   slots.CurrentSlot(uint64(s.genesisTime.Unix())),
			"idx":       index,
			"node":      node.ID(),
		}).Info("Prevote: filterPeerForAttSubnet: 444")

		return indExists
	}
}

// returns a method with filters peers specifically for a particular sync subnet.
func (s *Service) filterPeerForSyncSubnet(index uint64) func(node *enode.Node) bool {
	return func(node *enode.Node) bool {
		if !s.filterPeer(node) {
			return false
		}
		subnets, err := syncSubnets(node.Record())
		if err != nil {
			return false
		}
		indExists := false
		for _, comIdx := range subnets {
			if comIdx == index {
				indExists = true
				break
			}
		}
		return indExists
	}
}

// lower threshold to broadcast object compared to searching
// for a subnet. So that even in the event of poor peer
// connectivity, we can still broadcast an attestation.
func (s *Service) hasPeerWithSubnet(topic string) bool {
	// In the event peer threshold is lower, we will choose the lower
	// threshold.
	minPeers := mathutil.Min(1, uint64(flags.Get().MinimumPeersPerSubnet))
	return len(s.pubsub.ListPeers(topic+s.Encoding().ProtocolSuffix())) >= int(minPeers) // lint:ignore uintcast -- Min peers can be safely cast to int.
}

// Updates the service's discv5 listener record's attestation subnet
// with a new value for a bitfield of subnets tracked. It also updates
// the node's metadata by increasing the sequence number and the
// subnets tracked by the node.
func (s *Service) updateSubnetRecordWithMetadata(bitV bitfield.Bitvector64) {
	entry := enr.WithEntry(attSubnetEnrKey, &bitV)
	s.dv5Listener.LocalNode().Set(entry)
	s.metaData = wrapper.WrappedMetadataV0(&pb.MetaDataV0{
		SeqNumber: s.metaData.SequenceNumber() + 1,
		Attnets:   bitV,
	})
}

// Updates the service's discv5 listener record's attestation subnet
// with a new value for a bitfield of subnets tracked. It also record's
// the sync committee subnet in the enr. It also updates the node's
// metadata by increasing the sequence number and the subnets tracked by the node.
func (s *Service) updateSubnetRecordWithMetadataV2(bitVAtt bitfield.Bitvector64, bitVSync bitfield.Bitvector4) {
	entry := enr.WithEntry(attSubnetEnrKey, &bitVAtt)
	subEntry := enr.WithEntry(syncCommsSubnetEnrKey, &bitVSync)
	s.dv5Listener.LocalNode().Set(entry)
	s.dv5Listener.LocalNode().Set(subEntry)
	s.metaData = wrapper.WrappedMetadataV1(&pb.MetaDataV1{
		SeqNumber: s.metaData.SequenceNumber() + 1,
		Attnets:   bitVAtt,
		Syncnets:  bitVSync,
	})
}

// Initializes a bitvector of attestation subnets beacon nodes is subscribed to
// and creates a new ENR entry with its default value.
func initializeAttSubnets(node *enode.LocalNode) *enode.LocalNode {
	bitV := bitfield.NewBitvector64()
	entry := enr.WithEntry(attSubnetEnrKey, bitV.Bytes())
	node.Set(entry)
	return node
}

// Initializes a bitvector of sync committees subnets beacon nodes is subscribed to
// and creates a new ENR entry with its default value.
func initializeSyncCommSubnets(node *enode.LocalNode) *enode.LocalNode {
	bitV := bitfield.Bitvector4{byte(0x00)}
	entry := enr.WithEntry(syncCommsSubnetEnrKey, bitV.Bytes())
	node.Set(entry)
	return node
}

// Reads the attestation subnets entry from a node's ENR and determines
// the committee indices of the attestation subnets the node is subscribed to.
func attSubnets(record *enr.Record) ([]uint64, error) {
	bitV, err := attBitvector(record)
	if err != nil {
		return nil, err
	}
	// lint:ignore uintcast -- subnet count can be safely cast to int.
	if len(bitV) != byteCount(int(attestationSubnetCount)) {
		return []uint64{}, errors.Errorf("invalid bitvector provided, it has a size of %d", len(bitV))
	}
	var committeeIdxs []uint64
	for i := uint64(0); i < attestationSubnetCount; i++ {
		if bitV.BitAt(i) {
			committeeIdxs = append(committeeIdxs, i)
		}
	}
	return committeeIdxs, nil
}

// Reads the sync subnets entry from a node's ENR and determines
// the committee indices of the sync subnets the node is subscribed to.
func syncSubnets(record *enr.Record) ([]uint64, error) {
	bitV, err := syncBitvector(record)
	if err != nil {
		return nil, err
	}
	// lint:ignore uintcast -- subnet count can be safely cast to int.
	if len(bitV) != byteCount(int(syncCommsSubnetCount)) {
		return []uint64{}, errors.Errorf("invalid bitvector provided, it has a size of %d", len(bitV))
	}
	var committeeIdxs []uint64
	for i := uint64(0); i < syncCommsSubnetCount; i++ {
		if bitV.BitAt(i) {
			committeeIdxs = append(committeeIdxs, i)
		}
	}
	return committeeIdxs, nil
}

// Parses the attestation subnets ENR entry in a node and extracts its value
// as a bitvector for further manipulation.
func attBitvector(record *enr.Record) (bitfield.Bitvector64, error) {
	bitV := bitfield.NewBitvector64()
	entry := enr.WithEntry(attSubnetEnrKey, &bitV)
	err := record.Load(entry)
	if err != nil {
		return nil, err
	}
	return bitV, nil
}

// Parses the attestation subnets ENR entry in a node and extracts its value
// as a bitvector for further manipulation.
func syncBitvector(record *enr.Record) (bitfield.Bitvector4, error) {
	bitV := bitfield.Bitvector4{byte(0x00)}
	entry := enr.WithEntry(syncCommsSubnetEnrKey, &bitV)
	err := record.Load(entry)
	if err != nil {
		return nil, err
	}
	return bitV, nil
}

// The subnet locker is a map which keeps track of all
// mutexes stored per subnet. This locker is re-used
// between both the attestation and sync subnets. In
// order to differentiate between attestation and sync
// subnets. Sync subnets are stored by (subnet+syncLockerVal). This
// is to prevent conflicts while allowing both subnets
// to use a single locker.
func (s *Service) subnetLocker(i uint64) *sync.RWMutex {
	s.subnetsLockLock.Lock()
	defer s.subnetsLockLock.Unlock()
	l, ok := s.subnetsLock[i]
	if !ok {
		l = &sync.RWMutex{}
		s.subnetsLock[i] = l
	}
	return l
}

// Determines the number of bytes that are used
// to represent the provided number of bits.
func byteCount(bitCount int) int {
	numOfBytes := bitCount / 8
	if bitCount%8 != 0 {
		numOfBytes++
	}
	return numOfBytes
}
