package kv

import (
	"context"
	"errors"
	"reflect"

	fastssz "github.com/ferranbt/fastssz"
	"github.com/golang/snappy"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
	"google.golang.org/protobuf/proto"
)

func decode(ctx context.Context, data []byte, dst proto.Message) error {
	_, span := trace.StartSpan(ctx, "BeaconDB.decode")
	defer span.End()

	data, err := snappy.Decode(nil, data)
	if err != nil {
		return err
	}
	if isSSZStorageFormat(dst) {
		return dst.(fastssz.Unmarshaler).UnmarshalSSZ(data)
	}
	return proto.Unmarshal(data, dst)
}

func encode(ctx context.Context, msg proto.Message) ([]byte, error) {
	_, span := trace.StartSpan(ctx, "BeaconDB.encode")
	defer span.End()

	if msg == nil || reflect.ValueOf(msg).IsNil() {
		return nil, errors.New("cannot encode nil message")
	}
	var enc []byte
	var err error
	if isSSZStorageFormat(msg) {
		enc, err = msg.(fastssz.Marshaler).MarshalSSZ()
		if err != nil {
			return nil, err
		}
	} else {
		enc, err = proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
	}
	return snappy.Encode(nil, enc), nil
}

// isSSZStorageFormat returns true if the object type should be saved in SSZ encoded format.
func isSSZStorageFormat(obj interface{}) bool {
	if _, ok := obj.(*ethpb.BeaconState); ok {
		return true
	}

	if _, ok := obj.(*ethpb.SignedBeaconBlock); ok {
		return true
	}

	if _, ok := obj.(*ethpb.SignedAggregateAttestationAndProof); ok {
		return true
	}

	if _, ok := obj.(*ethpb.BeaconBlock); ok {
		return true
	}

	if _, ok := obj.(*ethpb.Attestation); ok {
		return true
	}

	if _, ok := obj.(*ethpb.Deposit); ok {
		return true
	}

	if _, ok := obj.(*ethpb.AttesterSlashing); ok {
		return true
	}

	if _, ok := obj.(*ethpb.ProposerSlashing); ok {
		return true
	}

	if _, ok := obj.(*ethpb.VoluntaryExit); ok {
		return true
	}

	return false
}
