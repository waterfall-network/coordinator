package sync

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/async/event"
	blockfeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/block"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/operation"
	statefeed "gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/feed/state"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/db"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/attestations"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/prevote"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/slashings"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/synccommittee"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/operations/voluntaryexits"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/p2p"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/state/stategen"
)

type Option func(s *Service) error

func WithAttestationNotifier(notifier operation.Notifier) Option {
	return func(s *Service) error {
		s.cfg.attestationNotifier = notifier
		return nil
	}
}

func WithP2P(p2p p2p.P2P) Option {
	return func(s *Service) error {
		s.cfg.p2p = p2p
		return nil
	}
}

func WithDatabase(db db.NoHeadAccessDatabase) Option {
	return func(s *Service) error {
		s.cfg.beaconDB = db
		return nil
	}
}

func WithAttestationPool(attPool attestations.Pool) Option {
	return func(s *Service) error {
		s.cfg.attPool = attPool
		return nil
	}
}

func WithPrevotePool(pvPool prevote.Pool) Option {
	return func(s *Service) error {
		s.cfg.prevotePool = pvPool
		return nil
	}
}

func WithExitPool(exitPool voluntaryexits.PoolManager) Option {
	return func(s *Service) error {
		s.cfg.exitPool = exitPool
		return nil
	}
}

func WithSlashingPool(slashingPool slashings.PoolManager) Option {
	return func(s *Service) error {
		s.cfg.slashingPool = slashingPool
		return nil
	}
}

func WithSyncCommsPool(syncCommsPool synccommittee.Pool) Option {
	return func(s *Service) error {
		s.cfg.syncCommsPool = syncCommsPool
		return nil
	}
}

func WithChainService(chain blockchainService) Option {
	return func(s *Service) error {
		s.cfg.chain = chain
		return nil
	}
}

func WithInitialSync(initialSync Checker) Option {
	return func(s *Service) error {
		s.cfg.initialSync = initialSync
		return nil
	}
}

func WithStateNotifier(stateNotifier statefeed.Notifier) Option {
	return func(s *Service) error {
		s.cfg.stateNotifier = stateNotifier
		return nil
	}
}

func WithBlockNotifier(blockNotifier blockfeed.Notifier) Option {
	return func(s *Service) error {
		s.cfg.blockNotifier = blockNotifier
		return nil
	}
}

func WithOperationNotifier(operationNotifier operation.Notifier) Option {
	return func(s *Service) error {
		s.cfg.operationNotifier = operationNotifier
		return nil
	}
}

func WithStateGen(stateGen *stategen.State) Option {
	return func(s *Service) error {
		s.cfg.stateGen = stateGen
		return nil
	}
}

func WithSlasherAttestationsFeed(slasherAttestationsFeed *event.Feed) Option {
	return func(s *Service) error {
		s.cfg.slasherAttestationsFeed = slasherAttestationsFeed
		return nil
	}
}

func WithSlasherBlockHeadersFeed(slasherBlockHeadersFeed *event.Feed) Option {
	return func(s *Service) error {
		s.cfg.slasherBlockHeadersFeed = slasherBlockHeadersFeed
		return nil
	}
}
