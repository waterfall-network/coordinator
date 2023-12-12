package apimiddleware

type listKeystoresResponseJSON struct {
	Keystores []*keystoreJSON `json:"data"`
}

type keystoreJSON struct {
	ValidatingPubkey string `json:"validating_pubkey" hex:"true"`
	DerivationPath   string `json:"derivation_path"`
}

type importKeystoresRequestJSON struct {
	Keystores          []string `json:"keystores"`
	Passwords          []string `json:"passwords"`
	SlashingProtection string   `json:"slashing_protection"`
}

type importKeystoresResponseJSON struct {
	Statuses []*statusJSON `json:"data"`
}

type deleteKeystoresRequestJSON struct {
	PublicKeys []string `json:"pubkeys" hex:"true"`
}

type statusJSON struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type deleteKeystoresResponseJSON struct {
	Statuses           []*statusJSON `json:"data"`
	SlashingProtection string        `json:"slashing_protection"`
}

//remote keymanager api

type listRemoteKeysResponseJSON struct {
	Keystores []*remoteKeysListJSON `json:"data"`
}

type remoteKeysListJSON struct {
	Pubkey   string `json:"pubkey" hex:"true"`
	URL      string `json:"url"`
	Readonly bool   `json:"readonly"`
}

type remoteKeysJSON struct {
	Pubkey   string `json:"pubkey" hex:"true"`
	Url      string `json:"url"`
	Readonly bool   `json:"readonly"`
}

type importRemoteKeysRequestJSON struct {
	Keystores []*remoteKeysJSON `json:"remote_keys"`
}

type importRemoteKeysResponseJSON struct {
	Statuses []*statusJSON `json:"data"`
}

type deleteRemoteKeysRequestJSON struct {
	PublicKeys []string `json:"pubkeys" hex:"true"`
}

type deleteRemoteKeysResponseJSON struct {
	Statuses []*statusJSON `json:"data"`
}
