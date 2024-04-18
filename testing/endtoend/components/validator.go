package components

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pkg/errors"
	cmdshared "gitlab.waterfall.network/waterfall/protocol/coordinator/cmd"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/validator/flags"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/features"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/interop"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/endtoend/helpers"
	e2e "gitlab.waterfall.network/waterfall/protocol/coordinator/testing/endtoend/params"
	e2etypes "gitlab.waterfall.network/waterfall/protocol/coordinator/testing/endtoend/types"
)

const depositGasLimit = 4000000

var _ e2etypes.ComponentRunner = (*ValidatorNode)(nil)
var _ e2etypes.ComponentRunner = (*ValidatorNodeSet)(nil)

// ValidatorNodeSet represents set of validator nodes.
type ValidatorNodeSet struct {
	e2etypes.ComponentRunner
	config  *e2etypes.E2EConfig
	started chan struct{}
}

// NewValidatorNodeSet creates and returns a set of validator nodes.
func NewValidatorNodeSet(config *e2etypes.E2EConfig) *ValidatorNodeSet {
	return &ValidatorNodeSet{
		config:  config,
		started: make(chan struct{}, 1),
	}
}

// Start starts the configured amount of validators, also sending and mining their deposits.
func (s *ValidatorNodeSet) Start(ctx context.Context) error {
	// Always using genesis count since using anything else would be difficult to test for.
	validatorNum := int(params.BeaconConfig().MinGenesisActiveValidatorCount)
	prysmBeaconNodeNum := e2e.TestParams.BeaconNodeCount
	beaconNodeNum := prysmBeaconNodeNum + e2e.TestParams.LighthouseBeaconNodeCount
	if validatorNum%beaconNodeNum != 0 {
		return errors.New("validator count is not easily divisible by beacon node count")
	}
	validatorsPerNode := validatorNum / beaconNodeNum

	// Create validator nodes.
	nodes := make([]e2etypes.ComponentRunner, prysmBeaconNodeNum)
	for i := 0; i < prysmBeaconNodeNum; i++ {
		nodes[i] = NewValidatorNode(s.config, validatorsPerNode, i, validatorsPerNode*i)
	}

	// Wait for all nodes to finish their job (blocking).
	// Once nodes are ready passed in handler function will be called.
	return helpers.WaitOnNodes(ctx, nodes, func() {
		// All nodes stated, close channel, so that all services waiting on a set, can proceed.
		close(s.started)
	})
}

// Started checks whether validator node set is started and all nodes are ready to be queried.
func (s *ValidatorNodeSet) Started() <-chan struct{} {
	return s.started
}

// ValidatorNode represents a validator node.
type ValidatorNode struct {
	e2etypes.ComponentRunner
	config       *e2etypes.E2EConfig
	started      chan struct{}
	validatorNum int
	index        int
	offset       int
}

// NewValidatorNode creates and returns a validator node.
func NewValidatorNode(config *e2etypes.E2EConfig, validatorNum, index, offset int) *ValidatorNode {
	return &ValidatorNode{
		config:       config,
		validatorNum: validatorNum,
		index:        index,
		offset:       offset,
		started:      make(chan struct{}, 1),
	}
}

// Start starts a validator client.
func (v *ValidatorNode) Start(ctx context.Context) error {
	var pkg, target string
	if v.config.UsePrysmShValidator {
		pkg = ""
		target = "prysm_sh"
	} else {
		pkg = "cmd/validator"
		target = "validator"
	}
	binaryPath, found := bazel.FindBinary(pkg, target)
	if !found {
		return errors.New("validator binary not found")
	}

	config, validatorNum, index, offset := v.config, v.validatorNum, v.index, v.offset
	beaconRPCPort := e2e.TestParams.Ports.PrysmBeaconNodeRPCPort + index
	if beaconRPCPort >= e2e.TestParams.Ports.PrysmBeaconNodeRPCPort+e2e.TestParams.BeaconNodeCount {
		// Point any extra validator clients to a node we know is running.
		beaconRPCPort = e2e.TestParams.Ports.PrysmBeaconNodeRPCPort
	}

	file, err := helpers.DeleteAndCreateFile(e2e.TestParams.LogPath, fmt.Sprintf(e2e.ValidatorLogFileName, index))
	if err != nil {
		return err
	}
	gFile, err := helpers.GraffitiYamlFile(e2e.TestParams.TestPath)
	if err != nil {
		return err
	}
	args := []string{
		fmt.Sprintf("--%s=%s/eth2-val-%d", cmdshared.DataDirFlag.Name, e2e.TestParams.TestPath, index),
		fmt.Sprintf("--%s=%s", cmdshared.LogFileName.Name, file.Name()),
		fmt.Sprintf("--%s=%s", flags.GraffitiFileFlag.Name, gFile),
		fmt.Sprintf("--%s=%d", flags.MonitoringPortFlag.Name, e2e.TestParams.Ports.ValidatorMetricsPort+index),
		fmt.Sprintf("--%s=%d", flags.GRPCGatewayPort.Name, e2e.TestParams.Ports.ValidatorGatewayPort+index),
		fmt.Sprintf("--%s=localhost:%d", flags.BeaconRPCProviderFlag.Name, beaconRPCPort),
		fmt.Sprintf("--%s=%s", flags.GrpcHeadersFlag.Name, "dummy=value,foo=bar"), // Sending random headers shouldn't break anything.
		fmt.Sprintf("--%s=%s", cmdshared.VerbosityFlag.Name, "debug"),
		"--" + cmdshared.ForceClearDB.Name,
		"--" + cmdshared.E2EConfigFlag.Name,
		"--" + cmdshared.AcceptTosFlag.Name,
	}
	// Only apply e2e flags to the current branch. New flags may not exist in previous release.
	if !v.config.UsePrysmShValidator {
		args = append(args, features.E2EValidatorFlags...)
	}
	if v.config.UseWeb3RemoteSigner {
		args = append(args, fmt.Sprintf("--%s=http://localhost:%d", flags.Web3SignerURLFlag.Name, Web3RemoteSignerPort))
		// Write the pubkeys as comma seperated hex strings with 0x prefix.
		// See: https://docs.teku.consensys.net/en/latest/HowTo/External-Signer/Use-External-Signer/
		_, pubs, err := interop.DeterministicallyGenerateKeys(uint64(offset), uint64(validatorNum))
		if err != nil {
			return err
		}
		var hexPubs []string
		for _, pub := range pubs {
			hexPubs = append(hexPubs, "0x"+hex.EncodeToString(pub.Marshal()))
		}
		args = append(args, fmt.Sprintf("--%s=%s", flags.Web3SignerPublicValidatorKeysFlag.Name, strings.Join(hexPubs, ",")))
	} else {
		// When not using remote key signer, use interop keys.
		args = append(args,
			fmt.Sprintf("--%s=%d", flags.InteropNumValidators.Name, validatorNum),
			fmt.Sprintf("--%s=%d", flags.InteropStartIndex.Name, offset))
	}
	args = append(args, config.ValidatorFlags...)

	if v.config.UsePrysmShValidator {
		args = append([]string{"validator"}, args...)
		log.Warning("Using latest release validator via prysm.sh")
	}

	cmd := exec.CommandContext(ctx, binaryPath, args...) // #nosec G204 -- Safe

	// Write stdout and stderr to log files.
	stdout, err := os.Create(path.Join(e2e.TestParams.LogPath, fmt.Sprintf("validator_%d_stdout.log", index)))
	if err != nil {
		return err
	}
	stderr, err := os.Create(path.Join(e2e.TestParams.LogPath, fmt.Sprintf("validator_%d_stderr.log", index)))
	if err != nil {
		return err
	}
	defer func() {
		if err := stdout.Close(); err != nil {
			log.WithError(err).Error("Failed to close stdout file")
		}
		if err := stderr.Close(); err != nil {
			log.WithError(err).Error("Failed to close stderr file")
		}
	}()
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	log.Infof("Starting validator client %d with flags: %s %s", index, binaryPath, strings.Join(args, " "))
	if err = cmd.Start(); err != nil {
		return err
	}

	// Mark node as ready.
	close(v.started)

	return cmd.Wait()
}

// Started checks whether validator node is started and ready to be queried.
func (v *ValidatorNode) Started() <-chan struct{} {
	return v.started
}
