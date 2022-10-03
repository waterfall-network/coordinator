package keymanager_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/waterfall-foundation/coordinator/testing/assert"
	"github.com/waterfall-foundation/coordinator/testing/require"
	"github.com/waterfall-foundation/coordinator/validator/keymanager"
	"github.com/waterfall-foundation/coordinator/validator/keymanager/derived"
	"github.com/waterfall-foundation/coordinator/validator/keymanager/local"
	"github.com/waterfall-foundation/coordinator/validator/keymanager/remote"
	remote_web3signer "github.com/waterfall-foundation/coordinator/validator/keymanager/remote-web3signer"
)

var (
	_ = keymanager.IKeymanager(&local.Keymanager{})
	_ = keymanager.IKeymanager(&derived.Keymanager{})
	_ = keymanager.IKeymanager(&remote.Keymanager{})

	// More granular assertions.
	_ = keymanager.KeysFetcher(&local.Keymanager{})
	_ = keymanager.KeysFetcher(&derived.Keymanager{})
	_ = keymanager.Importer(&local.Keymanager{})
	_ = keymanager.Importer(&derived.Keymanager{})
	_ = keymanager.Deleter(&local.Keymanager{})
	_ = keymanager.Deleter(&derived.Keymanager{})

	_ = keymanager.PublicKeyAdder(&remote_web3signer.Keymanager{})
	_ = keymanager.PublicKeyDeleter(&remote_web3signer.Keymanager{})
)

func TestKeystoreContainsPath(t *testing.T) {
	keystore := keymanager.Keystore{}
	encoded, err := json.Marshal(keystore)

	require.NoError(t, err, "Unexpected error marshalling keystore")
	assert.Equal(t, true, strings.Contains(string(encoded), "path"))
}
