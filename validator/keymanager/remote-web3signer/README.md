# Web3Signer

Web3Signer is a popular remote signer tool by Consensys to allow users to store validation keys outside the validation
client and signed without the vc knowing the private keys. Web3Signer Specs are found by
searching `Consensys' Web3Signer API specification`

issue: https://gitlab.waterfall.network/waterfall/protocol/coordinator/issues/9994

## Support

WIP

## Features

### CLI

WIP

### API

- Get Public keys: returns all public keys currently stored with web3signer excluding newly added keys if reload keys
  was not run.
- Sign: Signs a message with a given public key. There are several types of messages that can be signed ( web3signer
  type to prysm type):
    - BLOCK <- *validatorpb.SignRequest_Block
    - ATTESTATION <- *validatorpb.SignRequest_AttestationData
    - AGGREGATE_AND_PROOF <- *validatorpb.SignRequest_AggregateAttestationAndProof
    - AGGREGATION_SLOT <- *validatorpb.SignRequest_Slot
    - BLOCK_V2 <- *validatorpb.SignRequest_BlockV2
    - BLOCK_V3 <- *validatorpb.SignRequest_BlockV3
    - DEPOSIT <- not supported
    - RANDAO_REVEAL <- *validatorpb.SignRequest_Epoch
    - VOLUNTARY_EXIT <- *validatorpb.SignRequest_Exit
    - SYNC_COMMITTEE_MESSAGE <- *validatorpb.SignRequest_SyncMessageBlockRoot
    - SYNC_COMMITTEE_SELECTION_PROOF <- *validatorpb.SignRequest_SyncAggregatorSelectionData
    - SYNC_COMMITTEE_CONTRIBUTION_AND_PROOF <- *validatorpb.SignRequest_ContributionAndProof
- Reload Keys: reloads all public keys from the web3signer.
- Get Server Status: returns OK if the web3signer is ok.

## Files Added and Files Changed

- Files Added:
    - validator/keymanager/remote-web3signer package

- Files Modified:
    - modified:   cmd/validator/flags/flags.go
    - modified:   validator/accounts/accounts_backup.go
    - modified:   validator/accounts/accounts_list.go
    - modified:   validator/accounts/iface/wallet.go
    - modified:   validator/accounts/userprompt/prompt.go
    - modified:   validator/accounts/wallet/wallet.go
    - modified:   validator/accounts/wallet_create.go
    - modified:   validator/client/runner.go
    - modified:   validator/client/validator.go
    - modified:   validator/keymanager/remote-web3signer/keymanager.go
    - modified:   validator/keymanager/types.go