package rpc

import (
	pb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1/validator-client"
)

var _ pb.AuthServer = (*Server)(nil)
