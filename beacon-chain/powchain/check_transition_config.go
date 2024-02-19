package powchain

import (
	"context"
	"errors"
	"time"

	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common/hexutil"
)

var (
	checkTransitionPollingInterval = time.Second * 10
	configMismatchLog              = "Configuration mismatch between your execution client and Prysm. " +
		"Please check your execution client and restart it with the proper configuration. If this is not done, " +
		"your node will not be able to complete the proof-of-stake transition"
)

// We check if there is a configuration mismatch error between the execution client
// and the Prysm beacon node. If so, we need to log errors in the node as it cannot successfully
// complete the merge transition for the Bellatrix hard fork.
func (s *Service) handleExchangeConfigurationError(err error) {
	if err == nil {
		// If there is no error in checking the exchange configuration error, we clear
		// the run error of the service if we had previously set it to ErrConfigMismatch.
		if errors.Is(s.runError, ErrConfigMismatch) {
			s.runError = nil
		}
		return
	}
	// If the error is a configuration mismatch, we set a runtime error in the service.
	if errors.Is(err, ErrConfigMismatch) {
		s.runError = err
		log.WithError(err).Error(configMismatchLog)
		return
	}
	log.WithError(err).Error("Could not check configuration values between execution and consensus client")
}

// Logs the terminal total difficulty status.
func (s *Service) logTtdStatus(ctx context.Context, ttd *uint256.Int) (bool, error) {
	latest, err := s.LatestExecutionBlock(ctx)
	switch {
	case errors.Is(err, hexutil.ErrEmptyString):
		return false, nil
	case err != nil:
		return false, err
	case latest == nil:
		return false, errors.New("latest block is nil")
	case latest.TotalDifficulty == "":
		return false, nil
	default:
	}
	latestTtd, err := hexutil.DecodeBig(latest.TotalDifficulty)
	if err != nil {
		return false, err
	}
	if latestTtd.Cmp(ttd.ToBig()) >= 0 {
		return true, nil
	}
	log.WithFields(logrus.Fields{
		"latestDifficulty":   latestTtd.String(),
		"terminalDifficulty": ttd.ToBig().String(),
	}).Info("terminal difficulty has not been reached yet")
	return false, nil
}
