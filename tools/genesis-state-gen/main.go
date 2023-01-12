package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/io/file"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/runtime/interop"
	"gitlab.waterfall.network/waterfall/protocol/gwat/common"
)

// DepositDataJSON representing a json object of hex string and uint64 values for
// validators on Ethereum. This file can be generated using the official eth2.0-deposit-cli.
type DepositDataJSON struct {
	PubKey                string `json:"pubkey"`
	Amount                uint64 `json:"amount"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
	DepositDataRoot       string `json:"deposit_data_root"`
	Signature             string `json:"signature"`
}

var (
	depositJSONFile = flag.String(
		"deposit-json-file",
		"",
		"Path to deposit_data.json file generated by the eth2.0-deposit-cli tool",
	)
	numValidators    = flag.Int("num-validators", 0, "Number of validators to deterministically generate in the generated genesis state")
	useMainnetConfig = flag.Bool("mainnet-config", false, "Select whether genesis state should be generated with mainnet or minimal (default) params")
	genesisTime      = flag.Uint64("genesis-time", 0, "Unix timestamp used as the genesis time in the generated genesis state (defaults to now+120sec)")
	gwatGenesisHash  = flag.String("gwat-genesis-hash", "", "Hash of the GWAT genesis block")
	sszOutputFile    = flag.String("output-ssz", "", "Output filename of the SSZ marshaling of the generated genesis state")
	yamlOutputFile   = flag.String("output-yaml", "", "Output filename of the YAML marshaling of the generated genesis state")
	jsonOutputFile   = flag.String("output-json", "", "Output filename of the JSON marshaling of the generated genesis state")
)

func main() {
	flag.Parse()
	if len(*gwatGenesisHash) != 66 {
		log.Print("Bad value of --gwat-genesis-hash specified")
		return
	}
	gwatHash := common.HexToHash(*gwatGenesisHash)

	if *genesisTime == 0 {
		log.Print("No --genesis-time specified, defaulting to now + 2min")
	}
	if *sszOutputFile == "" && *yamlOutputFile == "" && *jsonOutputFile == "" {
		log.Println("Expected --output-ssz, --output-yaml, or --output-json to have been provided, received nil")
		return
	}
	// Note: generated genesis with minimal config doesn't work
	if !*useMainnetConfig {
		//params.OverrideBeaconConfig(params.MinimalSpecConfig())
	}
	var genesisState *ethpb.BeaconState
	var err error
	if *depositJSONFile != "" {
		inputFile := *depositJSONFile
		expanded, err := file.ExpandPath(inputFile)
		if err != nil {
			log.Printf("Could not expand file path %s: %v", inputFile, err)
			return
		}
		inputJSON, err := os.Open(expanded) // #nosec G304
		if err != nil {
			log.Printf("Could not open JSON file for reading: %v", err)
			return
		}
		defer func() {
			if err := inputJSON.Close(); err != nil {
				log.Printf("Could not close file %s: %v", inputFile, err)
			}
		}()
		log.Printf("Generating genesis state from input JSON deposit data %s", inputFile)
		genesisState, err = genesisStateFromJSONValidators(inputJSON, gwatHash, *genesisTime)
		if err != nil {
			log.Printf("Could not generate genesis beacon state: %v", err)
			return
		}
	} else {
		if *numValidators == 0 {
			log.Println("Expected --num-validators to have been provided, received 0")
			return
		}
		// If no JSON input is specified, we create the state deterministically from interop keys.
		genesisState, _, err = interop.GenerateGenesisState(context.Background(), *genesisTime, uint64(*numValidators))
		if err != nil {
			log.Printf("Could not generate genesis beacon state: %v", err)
			return
		}
	}

	genTime := new(big.Int).SetUint64(genesisState.GetGenesisTime())
	tm := time.Unix(genTime.Int64(), 0)
	log.Printf("set genesis time: %v  (timestamp=%d)", tm, genTime.Int64())

	if *sszOutputFile != "" {
		encodedState, err := genesisState.MarshalSSZ()
		if err != nil {
			log.Printf("Could not ssz marshal the genesis beacon state: %v", err)
			return
		}
		if err := file.WriteFile(*sszOutputFile, encodedState); err != nil {
			log.Printf("Could not write encoded genesis beacon state to file: %v", err)
			return
		}
		log.Printf("Done writing to %s", *sszOutputFile)
	}
	if *yamlOutputFile != "" {
		encodedState, err := yaml.Marshal(genesisState)
		if err != nil {
			log.Printf("Could not yaml marshal the genesis beacon state: %v", err)
			return
		}
		if err := file.WriteFile(*yamlOutputFile, encodedState); err != nil {
			log.Printf("Could not write encoded genesis beacon state to file: %v", err)
			return
		}
		log.Printf("Done writing to %s", *yamlOutputFile)
	}
	if *jsonOutputFile != "" {
		encodedState, err := json.Marshal(genesisState)
		if err != nil {
			log.Printf("Could not json marshal the genesis beacon state: %v", err)
			return
		}
		if err := file.WriteFile(*jsonOutputFile, encodedState); err != nil {
			log.Printf("Could not write encoded genesis beacon state to file: %v", err)
			return
		}
		log.Printf("Done writing to %s", *jsonOutputFile)
	}
}

func genesisStateFromJSONValidators(r io.Reader, gwtGenesisHash common.Hash, genesisTime uint64) (*ethpb.BeaconState, error) {
	enc, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var depositJSON []*DepositDataJSON
	if err := json.Unmarshal(enc, &depositJSON); err != nil {
		return nil, err
	}
	depositDataList := make([]*ethpb.Deposit_Data, len(depositJSON))
	depositDataRoots := make([][]byte, len(depositJSON))
	for i, val := range depositJSON {
		data, dataRootBytes, err := depositJSONToDepositData(val)
		if err != nil {
			return nil, err
		}
		depositDataList[i] = data
		depositDataRoots[i] = dataRootBytes
	}
	beaconState, _, err := interop.GenerateGenesisStateFromDepositData(context.Background(), gwtGenesisHash, genesisTime, depositDataList, depositDataRoots)
	if err != nil {
		return nil, err
	}
	return beaconState, nil
}

func depositJSONToDepositData(input *DepositDataJSON) (depositData *ethpb.Deposit_Data, dataRoot []byte, err error) {
	pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(input.PubKey, "0x"))
	if err != nil {
		return
	}
	withdrawalbytes, err := hex.DecodeString(strings.TrimPrefix(input.WithdrawalCredentials, "0x"))
	if err != nil {
		return
	}
	signatureBytes, err := hex.DecodeString(strings.TrimPrefix(input.Signature, "0x"))
	if err != nil {
		return
	}
	dataRootBytes, err := hex.DecodeString(strings.TrimPrefix(input.DepositDataRoot, "0x"))
	if err != nil {
		return
	}
	depositData = &ethpb.Deposit_Data{
		PublicKey:             pubKeyBytes,
		WithdrawalCredentials: withdrawalbytes,
		Amount:                input.Amount,
		Signature:             signatureBytes,
	}
	dataRoot = dataRootBytes
	return
}
