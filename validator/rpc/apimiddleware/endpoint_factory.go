package apimiddleware

import (
	"github.com/pkg/errors"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/api/gateway/apimiddleware"
)

// ValidatorEndpointFactory creates endpoints used for running validator API calls through the API Middleware.
type ValidatorEndpointFactory struct {
}

func (f *ValidatorEndpointFactory) IsNil() bool {
	return f == nil
}

// Paths is a collection of all valid validator API paths.
func (*ValidatorEndpointFactory) Paths() []string {
	return []string{
		"/eth/v1/keystores",
		"/eth/v1/remotekeys",
	}
}

// Create returns a new endpoint for the provided API path.
func (*ValidatorEndpointFactory) Create(path string) (*apimiddleware.Endpoint, error) {
	endpoint := apimiddleware.DefaultEndpoint()
	switch path {
	case "/eth/v1/keystores":
		endpoint.GetResponse = &listKeystoresResponseJSON{}
		endpoint.PostRequest = &importKeystoresRequestJSON{}
		endpoint.PostResponse = &importKeystoresResponseJSON{}
		endpoint.DeleteRequest = &deleteKeystoresRequestJSON{}
		endpoint.DeleteResponse = &deleteKeystoresResponseJSON{}
	case "/eth/v1/remotekeys":
		endpoint.GetResponse = &listRemoteKeysResponseJSON{}
		endpoint.PostRequest = &importRemoteKeysRequestJSON{}
		endpoint.PostResponse = &importRemoteKeysResponseJSON{}
		endpoint.DeleteRequest = &deleteRemoteKeysRequestJSON{}
		endpoint.DeleteResponse = &deleteRemoteKeysResponseJSON{}
	default:
		return nil, errors.New("invalid path")
	}
	endpoint.Path = path
	return &endpoint, nil
}
