package iface

import (
	"context"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/validator/keymanager"
	remote_web3signer "gitlab.waterfall.network/waterfall/protocol/coordinator/validator/keymanager/remote-web3signer"
)

// InitKeymanagerConfig defines configuration options for initializing a keymanager.
type InitKeymanagerConfig struct {
	ListenForChanges bool
	Web3SignerConfig *remote_web3signer.SetupConfig
}

// Wallet defines a struct which has capabilities and knowledge of how
// to read and write important accounts-related files to the filesystem.
// Useful for keymanagers to have persistent capabilities for accounts on-disk.
type Wallet interface {
	// Methods to retrieve wallet and accounts metadata.
	AccountsDir() string
	Password() string
	// Read methods for important wallet and accounts-related files.
	ReadFileAtPath(ctx context.Context, filePath string, fileName string) ([]byte, error)
	// Write methods to persist important wallet and accounts-related files to disk.
	WriteFileAtPath(ctx context.Context, pathName string, fileName string, data []byte) error
	// Method for initializing a new keymanager.
	InitializeKeymanager(ctx context.Context, cfg InitKeymanagerConfig) (keymanager.IKeymanager, error)
}
