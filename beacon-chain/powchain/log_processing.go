package powchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed"
	statefeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/helpers"
	coreState "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/transition"
	v1 "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/v1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	prysmTime "gitlab.waterfall.network/waterfall/protocol/coordinator/time"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/time/slots"
	gwat "gitlab.waterfall.network/waterfall/protocol/gwat"
	gwatCommon "gitlab.waterfall.network/waterfall/protocol/gwat/common"
	gwatTypes "gitlab.waterfall.network/waterfall/protocol/gwat/core/types"
	gwatVal "gitlab.waterfall.network/waterfall/protocol/gwat/validator"
)

const eth1DataSavingInterval = 1000
const maxTolerableDifference = 50
const defaultEth1HeaderReqLimit = uint64(1000)
const depositlogRequestLimit = 10000
const additiveFactorMultiplier = 0.10
const multiplicativeDecreaseDivisor = 2

func tooMuchDataRequestedError(err error) bool {
	// this error is only infura specific (other providers might have different error messages)
	return err.Error() == "query returned more than 10000 results"
}

// Eth2GenesisPowchainInfo retrieves the genesis time and eth1 block number of the beacon chain
// from the deposit contract.
func (s *Service) Eth2GenesisPowchainInfo() (uint64, *big.Int) {
	return s.chainStartData.GenesisTime, big.NewInt(int64(s.chainStartData.GenesisBlock))
}

// ProcessETH1Block processes the logs from the provided eth1Block.
func (s *Service) ProcessETH1Block(ctx context.Context, blkNum uint64) error {
	query := gwat.FilterQuery{
		Addresses: []gwatCommon.Address{
			s.cfg.depositContractAddr,
		},
		FromBlock: new(big.Int).SetUint64(blkNum),
		ToBlock:   new(big.Int).SetUint64(blkNum),
	}
	logs, err := s.httpLogger.FilterLogs(ctx, query)
	if err != nil {
		return err
	}
	for _, filterLog := range logs {
		// ignore logs that are not of the required block number
		if filterLog.BlockNumber != blkNum {
			continue
		}
		if err := s.ProcessLog(ctx, filterLog); err != nil {
			return errors.Wrap(err, "could not process log")
		}
	}
	if !s.chainStartData.Chainstarted {
		if err := s.checkBlockNumberForChainStart(ctx, blkNum); err != nil {
			return err
		}
	}
	return nil
}

// ProcessLog is the main method which handles the processing of all
// logs from the deposit contract on the ETH1.0 chain.
func (s *Service) ProcessLog(ctx context.Context, depositLog gwatTypes.Log) error {
	s.processingLock.RLock()
	defer s.processingLock.RUnlock()
	// Process logs according to their event signature.
	if depositLog.Topics[0] == gwatVal.EvtDepositLogSignature {
		if err := s.ProcessDepositLog(ctx, depositLog); err != nil {
			return errors.Wrap(err, "Could not process deposit log")
		}
		if s.lastReceivedMerkleIndex%eth1DataSavingInterval == 0 {
			return s.savePowchainData(ctx)
		}
		return nil
	}
	if depositLog.Topics[0] == gwatVal.EvtExitReqLogSignature {
		if err := s.ProcessExitLog(ctx, depositLog); err != nil {
			return errors.Wrap(err, "Could not process exit log")
		}
		return nil
	}
	log.WithField("signature", fmt.Sprintf("%#x", depositLog.Topics[0])).Debug("Not a valid event signature")
	return nil
}

func (s *Service) ProcessExitLog(ctx context.Context, exitLog gwatTypes.Log) error {
	pubkey, creatorAddr, valIndex, exitEpoch, err := gwatVal.UnpackExitRequestLogData(exitLog.Data)

	log.WithError(err).WithFields(logrus.Fields{
		"valIndex":    valIndex,
		"exitEpoch":   exitEpoch,
		"pubkey":      fmt.Sprintf("%#x", pubkey),
		"creatorAddr": fmt.Sprintf("%#x", creatorAddr),
	}).Info("Processing exit")

	if err != nil {
		return errors.Wrap(err, "Could not unpack log")
	}

	totalSecondsPassed := uint64(prysmTime.Now().Unix()) - s.cfg.finalizedStateAtStartup.GenesisTime()
	currentEpoch := types.Epoch(uint64(totalSecondsPassed) / uint64(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot)))

	if exitEpoch != nil && *exitEpoch > uint64(currentEpoch) {
		currentEpoch = types.Epoch(*exitEpoch)
	}

	exit := &ethpb.VoluntaryExit{Epoch: currentEpoch, ValidatorIndex: types.ValidatorIndex(valIndex)}

	// add tx data as sig
	adddata := make([]byte, 0, 96)
	adddata = append(adddata, pubkey.Bytes()...)
	adddata = append(adddata, exitLog.TxHash.Bytes()...)
	adddata = append(adddata, bytesutil.ToBytes(exitLog.BlockNumber, 8)...)
	adddata = append(adddata, bytesutil.ToBytes(uint64(exitLog.TxIndex), 8)...)
	signedExit := &ethpb.SignedVoluntaryExit{Exit: exit, Signature: adddata}

	s.cfg.exitPool.InsertVoluntaryExitByGwat(s.ctx, signedExit)

	return nil
}

// ProcessDepositLog processes the log which had been received from
// the ETH1.0 chain by trying to ascertain which participant deposited
// in the contract.
func (s *Service) ProcessDepositLog(ctx context.Context, depositLog gwatTypes.Log) error {
	pubkey, creatorAddr, withdrawalCredentials, amount, signature, depositIndex, err := gwatVal.UnpackDepositLogData(depositLog.Data)

	log.WithError(err).WithFields(logrus.Fields{
		"amount":          amount,
		"pubkey":          fmt.Sprintf("%#x", pubkey),
		"creatorAddr":     creatorAddr.Hex(),
		"withdrawalCreds": withdrawalCredentials.Hex(),
		"depositIndex":    depositIndex,
		"s.lastIndex":     s.lastReceivedMerkleIndex,
	}).Info("Processing deposit")

	if err != nil {
		return errors.Wrap(err, "Could not unpack log")
	}
	// If we have already seen this Merkle index, skip processing the log.
	// This can happen sometimes when we receive the same log twice from the
	// ETH1.0 network, and prevents us from updating our trie
	// with the same log twice, causing an inconsistent state root.
	index := int64(depositIndex) // lint:ignore uintcast -- MerkleTreeIndex should not exceed int64 in your lifetime.
	if index <= s.lastReceivedMerkleIndex {
		return nil
	}

	if index != s.lastReceivedMerkleIndex+1 {
		missedDepositLogsCount.Inc()
		return errors.Errorf("received incorrect merkle index: wanted %d but got %d", s.lastReceivedMerkleIndex+1, index)
	}
	s.lastReceivedMerkleIndex = index

	// We then decode the deposit input in order to create a deposit object
	// we can store in our persistent DB.
	depositData := &ethpb.Deposit_Data{
		Amount:                amount,
		PublicKey:             pubkey.Bytes(),
		Signature:             signature.Bytes(),
		CreatorAddress:        creatorAddr.Bytes(),
		WithdrawalCredentials: withdrawalCredentials.Bytes(),
	}

	depositHash, err := depositData.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "Unable to determine hashed value of deposit")
	}

	// Defensive check to validate incoming index.
	if s.depositTrie.NumOfItems() != int(index) {
		return errors.Errorf("invalid deposit index received: wanted %d but got %d", s.depositTrie.NumOfItems(), index)
	}
	if err = s.depositTrie.Insert(depositHash[:], int(index)); err != nil {
		return err
	}

	deposit := &ethpb.Deposit{
		Data: depositData,
	}
	// Only generate the proofs during pre-genesis.
	if !s.chainStartData.Chainstarted {
		proof, err := s.depositTrie.MerkleProof(int(index))
		if err != nil {
			return errors.Wrap(err, "Unable to generate merkle proof for deposit")
		}
		deposit.Proof = proof
	}

	// We always store all historical deposits in the DB.
	err = s.cfg.depositCache.InsertDeposit(ctx, deposit, depositLog.BlockNumber, index, s.depositTrie.HashTreeRoot())
	if err != nil {
		return errors.Wrap(err, "unable to insert deposit into cache")
	}
	validData := true
	if !s.chainStartData.Chainstarted {
		s.chainStartData.ChainstartDeposits = append(s.chainStartData.ChainstartDeposits, deposit)
		root := s.depositTrie.HashTreeRoot()
		eth1Data := &ethpb.Eth1Data{
			DepositRoot:  root[:],
			DepositCount: uint64(len(s.chainStartData.ChainstartDeposits)),
		}
		if err := s.processDeposit(ctx, eth1Data, deposit); err != nil {
			log.Errorf("Invalid deposit processed: %v", err)
			validData = false
		}
	} else {
		s.cfg.depositCache.InsertPendingDeposit(ctx, deposit, depositLog.BlockNumber, index, s.depositTrie.HashTreeRoot())
	}
	if validData {
		log.WithFields(logrus.Fields{
			"eth1Block":    depositLog.BlockNumber,
			"publicKey":    fmt.Sprintf("%#x", depositData.PublicKey),
			"depositIndex": index,
		}).Info("Deposit registered from gwat deposit event")
		validDepositsCount.Inc()
		// Notify users what is going on, from time to time.
		if !s.chainStartData.Chainstarted {
			deposits := len(s.chainStartData.ChainstartDeposits)
			if deposits%512 == 0 {
				valCount, err := helpers.ActiveValidatorCount(ctx, s.preGenesisState, 0)
				if err != nil {
					log.WithError(err).Error("Could not determine active validator count from pre genesis state")
				}
				log.WithFields(logrus.Fields{
					"deposits":          deposits,
					"genesisValidators": valCount,
				}).Info("Processing deposits")
			}
		}
	} else {
		log.WithFields(logrus.Fields{
			"eth1Block":    depositLog.BlockHash.Hex(),
			"eth1Tx":       depositLog.TxHash.Hex(),
			"depositIndex": index,
		}).Info("Invalid deposit registered in gwat deposit event")
	}
	return nil
}

// ProcessChainStart processes the log which had been received from
// the ETH1.0 chain by trying to determine when to start the beacon chain.
func (s *Service) ProcessChainStart(genesisTime uint64, eth1BlockHash [32]byte, blockNumber uint64) {
	s.chainStartData.Chainstarted = true
	s.chainStartData.GenesisBlock = blockNumber

	chainStartTime := time.Unix(int64(genesisTime), 0) // lint:ignore uintcast -- Genesis time wont exceed int64 in your lifetime.

	for i := range s.chainStartData.ChainstartDeposits {
		proof, err := s.depositTrie.MerkleProof(i)
		if err != nil {
			log.Errorf("Unable to generate deposit proof %v", err)
		}
		s.chainStartData.ChainstartDeposits[i].Proof = proof
	}

	root := s.depositTrie.HashTreeRoot()

	s.chainStartData.Eth1Data = &ethpb.Eth1Data{
		DepositCount: uint64(len(s.chainStartData.ChainstartDeposits)),
		DepositRoot:  root[:],
		BlockHash:    eth1BlockHash[:],
		Candidates:   []byte{},
		Finalization: []byte{},
	}

	log.WithFields(logrus.Fields{
		"ChainStartTime": chainStartTime,
	}).Info("Minimum number of validators reached for beacon-chain to start")
	s.cfg.stateNotifier.StateFeed().Send(&feed.Event{
		Type: statefeed.ChainStarted,
		Data: &statefeed.ChainStartedData{
			StartTime: chainStartTime,
		},
	})
	if err := s.savePowchainData(s.ctx); err != nil {
		// continue on, if the save fails as this will get re-saved
		// in the next interval.
		log.Error(err)
	}
}

func createGenesisTime(timeStamp uint64) uint64 {
	// adds in the genesis delay to the eth1 block time
	// on which it was triggered.
	return timeStamp + params.BeaconConfig().GenesisDelay
}

// processPastLogs processes all the past logs from the deposit contract and
// updates the deposit trie with the data from each individual log.
func (s *Service) processPastLogs(ctx context.Context) error {
	currentBlockNum := s.latestEth1Data.LastRequestedBlock
	deploymentBlock := params.BeaconNetworkConfig().ContractDeploymentBlock
	// Start from the deployment block if our last requested block
	// is behind it. This is as the deposit logs can only start from the
	// block of the deployment of the deposit contract.
	if deploymentBlock > currentBlockNum {
		currentBlockNum = deploymentBlock
	}
	// To store all blocks.
	headersMap := make(map[uint64]*gwatTypes.Header)
	logCount, err := s.GetDepositCount(ctx)
	if err != nil {
		return err
	}

	// Batch request the desired headers and store them in a
	// map for quick access.
	requestHeaders := func(startBlk uint64, endBlk uint64) error {
		headers, err := s.batchRequestHeaders(startBlk, endBlk)
		if err != nil {
			return err
		}
		for _, h := range headers {
			if h != nil && h.Number != nil {
				headersMap[h.Nr()] = h
			}
		}
		return nil
	}
	latestFollowHeight, err := s.followBlockHeight(ctx)
	if err != nil {
		return err
	}

	batchSize := s.cfg.eth1HeaderReqLimit
	additiveFactor := uint64(float64(batchSize) * additiveFactorMultiplier)

	for currentBlockNum < latestFollowHeight {
		start := currentBlockNum
		end := currentBlockNum + batchSize
		// Appropriately bound the request, as we do not
		// want request blocks beyond the current follow distance.
		if end > latestFollowHeight {
			end = latestFollowHeight
		}
		query := gwat.FilterQuery{
			Addresses: []gwatCommon.Address{
				s.cfg.depositContractAddr,
			},
			FromBlock: big.NewInt(0).SetUint64(start),
			ToBlock:   big.NewInt(0).SetUint64(end),
		}
		remainingLogs := logCount - uint64(s.lastReceivedMerkleIndex+1)
		// only change the end block if the remaining logs are below the required log limit.
		// reset our query and end block in this case.
		withinLimit := remainingLogs < depositlogRequestLimit
		aboveFollowHeight := end >= latestFollowHeight
		if withinLimit && aboveFollowHeight {
			query.ToBlock = big.NewInt(0).SetUint64(latestFollowHeight)
			end = latestFollowHeight
		}
		logs, err := s.httpLogger.FilterLogs(ctx, query)
		if err != nil {
			if tooMuchDataRequestedError(err) {
				if batchSize == 0 {
					return errors.New("batch size is zero")
				}

				// multiplicative decrease
				batchSize /= multiplicativeDecreaseDivisor
				continue
			}
			return err
		}
		// Only request headers before chainstart to correctly determine
		// genesis.
		if !s.chainStartData.Chainstarted {
			if err := requestHeaders(start, end); err != nil {
				return err
			}
		}

		for _, filterLog := range logs {
			if filterLog.BlockNumber > currentBlockNum {
				if err := s.checkHeaderRange(ctx, currentBlockNum, filterLog.BlockNumber-1, headersMap, requestHeaders); err != nil {
					return err
				}
				// set new block number after checking for chainstart for previous block.
				s.latestEth1Data.LastRequestedBlock = currentBlockNum
				currentBlockNum = filterLog.BlockNumber
			}
			if err := s.ProcessLog(ctx, filterLog); err != nil {
				return err
			}
		}
		if err := s.checkHeaderRange(ctx, currentBlockNum, end, headersMap, requestHeaders); err != nil {
			return err
		}
		currentBlockNum = end

		if batchSize < s.cfg.eth1HeaderReqLimit {
			// update the batchSize with additive increase
			batchSize += additiveFactor
			if batchSize > s.cfg.eth1HeaderReqLimit {
				batchSize = s.cfg.eth1HeaderReqLimit
			}
		}
	}

	s.latestEth1Data.LastRequestedBlock = currentBlockNum

	c, err := s.cfg.beaconDB.FinalizedCheckpoint(ctx)
	if err != nil {
		return err
	}
	fRoot := bytesutil.ToBytes32(c.Root)
	// Return if no checkpoint exists yet.
	if fRoot == params.BeaconConfig().ZeroHash {
		return nil
	}
	fState := s.cfg.finalizedStateAtStartup
	isNil := fState == nil || fState.IsNil()

	// If processing past logs take a long time, we
	// need to check if this is the correct finalized
	// state we are referring to and whether our cached
	// finalized state is referring to our current finalized checkpoint.
	// The current code does ignore an edge case where the finalized
	// block is in a different epoch from the checkpoint's epoch.
	// This only happens in skipped slots, so pruning it is not an issue.
	if isNil || slots.ToEpoch(fState.Slot()) != c.Epoch {
		fState, err = s.cfg.stateGen.StateByRoot(ctx, fRoot)
		if err != nil {
			return err
		}
	}
	if fState != nil && !fState.IsNil() && fState.Eth1DepositIndex() > 0 {
		s.cfg.depositCache.PrunePendingDeposits(ctx, int64(fState.Eth1DepositIndex())) // lint:ignore uintcast -- Deposit index should not exceed int64 in your lifetime.
	}
	return nil
}

// requestBatchedHeadersAndLogs requests and processes all the headers and
// logs from the period last polled to now.
func (s *Service) requestBatchedHeadersAndLogs(ctx context.Context) error {
	// We request for the nth block behind the current head, in order to have
	// stabilized logs when we retrieve it from the 1.0 chain.

	requestedBlock, err := s.followBlockHeight(ctx)
	if err != nil {
		return err
	}
	if requestedBlock > s.latestEth1Data.LastRequestedBlock &&
		requestedBlock-s.latestEth1Data.LastRequestedBlock > maxTolerableDifference {
		log.Infof("Falling back to historical headers and logs sync. Current difference is %d", requestedBlock-s.latestEth1Data.LastRequestedBlock)
		return s.processPastLogs(ctx)
	}
	for i := s.latestEth1Data.LastRequestedBlock + 1; i <= requestedBlock; i++ {
		// Cache eth1 block header here.
		_, err := s.BlockHashByHeight(ctx, big.NewInt(0).SetUint64(i))
		if err != nil {
			return err
		}
		err = s.ProcessETH1Block(ctx, i)
		if err != nil {
			return err
		}
		s.latestEth1Data.LastRequestedBlock = i
	}

	return nil
}

func (s *Service) retrieveBlockHashAndTime(ctx context.Context, blkNum *big.Int) ([32]byte, uint64, error) {
	bHash, err := s.BlockHashByHeight(ctx, blkNum)
	if err != nil {
		return [32]byte{}, 0, errors.Wrap(err, "could not get eth1 block hash")
	}
	if bHash == [32]byte{} {
		return [32]byte{}, 0, errors.Wrap(err, "got empty block hash")
	}
	timeStamp, err := s.BlockTimeByHeight(ctx, blkNum)
	if err != nil {
		return [32]byte{}, 0, errors.Wrap(err, "could not get block timestamp")
	}
	return bHash, timeStamp, nil
}

// checkBlockNumberForChainStart checks the given block number for if chainstart has occurred.
func (s *Service) checkBlockNumberForChainStart(ctx context.Context, blkNum uint64) error {
	bHash, timeStamp, err := s.retrieveBlockHashAndTime(ctx, new(big.Int).SetUint64(blkNum))
	if err != nil {
		return err
	}
	s.checkForChainstart(ctx, bHash, blkNum, timeStamp)
	return nil
}

func (s *Service) checkHeaderForChainstart(ctx context.Context, header *gwatTypes.Header) {
	s.checkForChainstart(ctx, header.Hash(), header.Nr(), header.Time)
}

func (s *Service) checkHeaderRange(ctx context.Context, start, end uint64, headersMap map[uint64]*gwatTypes.Header,
	requestHeaders func(uint64, uint64) error) error {
	for i := start; i <= end; i++ {
		if !s.chainStartData.Chainstarted {
			h, ok := headersMap[i]
			if !ok {
				if err := requestHeaders(i, end); err != nil {
					return err
				}
				// Retry this block.
				i--
				continue
			}
			s.checkHeaderForChainstart(ctx, h)
		}
	}
	return nil
}

// retrieves the current active validator count and genesis time from
// the provided block time.
func (s *Service) currentCountAndTime(ctx context.Context, blockTime uint64) (uint64, uint64) {
	if s.preGenesisState.NumValidators() == 0 {
		return 0, 0
	}
	valCount, err := helpers.ActiveValidatorCount(ctx, s.preGenesisState, 0)
	if err != nil {
		log.WithError(err).Error("Could not determine active validator count from pre genesis state")
		return 0, 0
	}
	return valCount, createGenesisTime(blockTime)
}

func (s *Service) checkForChainstart(ctx context.Context, blockHash [32]byte, blockNumber uint64, blockTime uint64) {
	valCount, genesisTime := s.currentCountAndTime(ctx, blockTime)
	if valCount == 0 {
		return
	}
	triggered := coreState.IsValidGenesisState(valCount, genesisTime)
	if triggered {
		s.chainStartData.GenesisTime = genesisTime
		s.ProcessChainStart(s.chainStartData.GenesisTime, blockHash, blockNumber)
	}
}

// save all powchain related metadata to disk.
func (s *Service) savePowchainData(ctx context.Context) error {
	pbState, err := v1.ProtobufBeaconState(s.preGenesisState.InnerStateUnsafe())
	if err != nil {
		return err
	}
	eth1Data := &ethpb.ETH1ChainData{
		CurrentEth1Data:   s.latestEth1Data,
		ChainstartData:    s.chainStartData,
		BeaconState:       pbState, // I promise not to mutate it!
		Trie:              s.depositTrie.ToProto(),
		DepositContainers: s.cfg.depositCache.AllDepositContainers(ctx),
	}
	return s.cfg.beaconDB.SavePowchainData(ctx, eth1Data)
}
