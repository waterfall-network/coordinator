package accounts

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	grpcutil "gitlab.waterfall.network/waterfall/protocol/coordinator/api/grpc"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/beacon-chain/core/blocks"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/cmd/validator/flags"
	fieldparams "gitlab.waterfall.network/waterfall/protocol/coordinator/config/fieldparams"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/encoding/bytesutil"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/accounts/iface"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/accounts/wallet"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/client"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/keymanager"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common/hexutil"
	"google.golang.org/grpc"
)

// PerformExitCfg for account voluntary exits.
type PerformExitCfg struct {
	ValidatorClient  ethpb.BeaconNodeValidatorClient
	NodeClient       ethpb.NodeClient
	Keymanager       keymanager.IKeymanager
	RawPubKeys       [][]byte
	FormattedPubKeys []string
}

const exitPassphrase = "Exit my validator"

// PerformVoluntaryExit uses gRPC clients to submit a voluntary exit message to a beacon node.
func PerformVoluntaryExit(
	ctx context.Context, cfg PerformExitCfg,
) (rawExitedKeys [][]byte, formattedExitedKeys []string, err error) {
	var rawNotExitedKeys [][]byte
	for i, key := range cfg.RawPubKeys {
		if err := client.ProposeExit(ctx, cfg.ValidatorClient, cfg.NodeClient, cfg.Keymanager.Sign, key); err != nil {
			rawNotExitedKeys = append(rawNotExitedKeys, key)

			msg := err.Error()
			if strings.Contains(msg, blocks.ValidatorAlreadyExitedMsg) ||
				strings.Contains(msg, blocks.ValidatorCannotExitYetMsg) {
				log.Warningf("Could not perform voluntary exit for account %s: %s", cfg.FormattedPubKeys[i], msg)
			} else {
				log.WithError(err).Errorf("voluntary exit failed for account %s", cfg.FormattedPubKeys[i])
			}
		}
	}

	rawExitedKeys = make([][]byte, 0)
	formattedExitedKeys = make([]string, 0)
	for i, key := range cfg.RawPubKeys {
		found := false
		for _, notExited := range rawNotExitedKeys {
			if bytes.Equal(notExited, key) {
				found = true
				break
			}
		}
		if !found {
			rawExitedKeys = append(rawExitedKeys, key)
			formattedExitedKeys = append(formattedExitedKeys, cfg.FormattedPubKeys[i])
		}
	}

	return rawExitedKeys, formattedExitedKeys, nil
}

func prepareWallet(cliCtx *cli.Context) (validatingPublicKeys [][fieldparams.BLSPubkeyLength]byte, km keymanager.IKeymanager, err error) {
	w, err := wallet.OpenWalletOrElseCli(cliCtx, func(cliCtx *cli.Context) (*wallet.Wallet, error) {
		return nil, wallet.ErrNoWalletFound
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not open wallet")
	}
	// TODO(#9883) - Remove this when we have a better way to handle this.
	if w.KeymanagerKind() == keymanager.Web3Signer {
		return nil, nil, errors.New(
			"web3signer wallets cannot exit accounts through cli command yet. please perform this on the remote signer node",
		)
	}
	km, err = w.InitializeKeymanager(cliCtx.Context, iface.InitKeymanagerConfig{ListenForChanges: false})
	if err != nil {
		return nil, nil, errors.Wrap(err, ErrCouldNotInitializeKeymanager)
	}
	validatingPublicKeys, err = km.FetchValidatingPublicKeys(cliCtx.Context)
	if err != nil {
		return nil, nil, err
	}
	if len(validatingPublicKeys) == 0 {
		return nil, nil, errors.New("wallet is empty, no accounts to perform voluntary exit")
	}

	return validatingPublicKeys, km, nil
}

func prepareAllKeys(validatingKeys [][fieldparams.BLSPubkeyLength]byte) (raw [][]byte, formatted []string) {
	raw = make([][]byte, len(validatingKeys))
	formatted = make([]string, len(validatingKeys))
	for i, pk := range validatingKeys {
		raw[i] = make([]byte, len(pk))
		copy(raw[i], pk[:])
		formatted[i] = fmt.Sprintf("%#x", bytesutil.Trunc(pk[:]))
	}
	return
}

func prepareClients(cliCtx *cli.Context) (*ethpb.BeaconNodeValidatorClient, *ethpb.NodeClient, error) {
	dialOpts := client.ConstructDialOptions(
		cliCtx.Int(cmd.GrpcMaxCallRecvMsgSizeFlag.Name),
		cliCtx.String(flags.CertFlag.Name),
		cliCtx.Uint(flags.GrpcRetriesFlag.Name),
		cliCtx.Duration(flags.GrpcRetryDelayFlag.Name),
	)
	if dialOpts == nil {
		return nil, nil, errors.New("failed to construct dial options")
	}

	grpcHeaders := strings.Split(cliCtx.String(flags.GrpcHeadersFlag.Name), ",")
	cliCtx.Context = grpcutil.AppendHeaders(cliCtx.Context, grpcHeaders)

	conn, err := grpc.DialContext(cliCtx.Context, cliCtx.String(flags.BeaconRPCProviderFlag.Name), dialOpts...)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not dial endpoint %s", flags.BeaconRPCProviderFlag.Name)
	}
	validatorClient := ethpb.NewBeaconNodeValidatorClient(conn)
	nodeClient := ethpb.NewNodeClient(conn)
	return &validatorClient, &nodeClient, nil
}

func displayExitInfo(rawExitedKeys [][]byte, trimmedExitedKeys []string) {
	if len(rawExitedKeys) > 0 {
		urlFormattedPubKeys := make([]string, len(rawExitedKeys))
		for i, key := range rawExitedKeys {
			var baseUrl string
			if params.BeaconConfig().ConfigName == params.ConfigNames[params.Pyrmont] {
				baseUrl = "https://pyrmont.beaconcha.in/validator/"
			} else if params.BeaconConfig().ConfigName == params.ConfigNames[params.Testnet8] {
				baseUrl = "https://testnet8.beaconcha.in/validator/"
			} else if params.BeaconConfig().ConfigName == params.ConfigNames[params.Testnet5] {
				baseUrl = "https://testnet5.beaconcha.in/validator/"
			} else if params.BeaconConfig().ConfigName == params.ConfigNames[params.Testnet9] {
				baseUrl = "https://testnet9.beaconcha.in/validator/"
			} else if params.BeaconConfig().ConfigName == params.ConfigNames[params.Mainnet] {
				baseUrl = "https://mainnet.beaconcha.in/validator/"
			} else {
				baseUrl = "https://beaconcha.in/validator/"
			}
			// Remove '0x' prefix
			urlFormattedPubKeys[i] = baseUrl + hexutil.Encode(key)[2:]
		}

		ifaceKeys := make([]interface{}, len(urlFormattedPubKeys))
		for i, k := range urlFormattedPubKeys {
			ifaceKeys[i] = k
		}

		info := fmt.Sprintf("Voluntary exit was successful for the accounts listed. "+
			"URLs where you can track each validator's exit:\n"+strings.Repeat("%s\n", len(ifaceKeys)), ifaceKeys...)

		log.WithField("publicKeys", strings.Join(trimmedExitedKeys, ", ")).Info(info)
	} else {
		log.Info("No successful voluntary exits")
	}
}
