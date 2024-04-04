package ssz_static_test

import (
	"fmt"
	"testing"

	fssz "github.com/ferranbt/fastssz"
	"github.com/pkg/errors"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	common "gitlab.waterfall.network/waterfall/protocol/coordinator/testing/spectest/shared/common/ssz_static"
)

func ExampleRunSSZStaticTests() {
	// Define an unmarshaller to select the correct go type based on the string
	// name provided in spectests and then populate it with the serialized bytes.
	unmarshaller := func(t *testing.T, serializedBytes []byte, objectName string) (interface{}, error) {
		var obj interface{}
		switch objectName {
		case "Attestation":
			obj = &ethpb.Attestation{}
		case "BeaconState":
			obj = &ethpb.BeaconState{}
		case "Eth1Block":
			// Some types may not apply to prysm, but exist in the spec test folders. It is OK to
			// skip these tests with a valid justification. Otherwise, the test should fail with an
			// unsupported type.
			t.Skip("Unused type")
			return nil, nil
		default:
			return nil, fmt.Errorf("unsupported type: %s", objectName)
		}
		var err error
		if o, ok := obj.(fssz.Unmarshaler); ok {
			err = o.UnmarshalSSZ(serializedBytes)
		} else {
			err = errors.New("could not unmarshal object, not a fastssz compatible object")
		}
		return obj, err
	}

	// Optional: define a method to add custom HTR methods for a given object.
	// This argument may be nil if your test does not require custom HTR methods.
	// Most commonly, this is used when a handwritten HTR method with specialized caching
	// is used and you want to ensure it passes spectests.
	var t *testing.T
	// common.RunSSZStaticTests will run all of the tests found in the spec test folder with the
	// given config and forkOrPhase. It will then use the unmarshaller to hydrate the types and
	// ensure that fastssz generated methods match the expected results. It will also test custom
	// HTR methods if provided.
	common.RunSSZStaticTests(t,
		"mainnet",    // Network configuration
		"phase0",     // Fork or phase
		unmarshaller) // nil customHTR is acceptable.
}
