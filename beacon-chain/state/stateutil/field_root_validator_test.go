package stateutil

import (
	"reflect"
	"strings"
	"testing"

	mathutil "gitlab.waterfall.network/waterfall/protocol/coordinator/math"
	ethpb "gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
)

func TestValidatorConstants(t *testing.T) {
	v := &ethpb.Validator{}
	refV := reflect.ValueOf(v).Elem()
	numFields := refV.NumField()
	numOfValFields := 0

	for i := 0; i < numFields; i++ {
		if strings.Contains(refV.Type().Field(i).Name, "state") ||
			strings.Contains(refV.Type().Field(i).Name, "sizeCache") ||
			strings.Contains(refV.Type().Field(i).Name, "unknownFields") ||
			strings.Contains(refV.Type().Field(i).Name, "PublicKey") ||
			strings.Contains(refV.Type().Field(i).Name, "CreatorAddress") ||
			strings.Contains(refV.Type().Field(i).Name, "WithdrawalCredentials") ||
			strings.Contains(refV.Type().Field(i).Name, "EffectiveBalance") {
			continue
		}
		numOfValFields++
	}
	assert.Equal(t, validatorFieldRoots, numOfValFields)
	assert.Equal(t, uint64(validatorFieldRoots), mathutil.PowerOf2(validatorTreeDepth))

	_, err := ValidatorRegistryRoot([]*ethpb.Validator{v})
	assert.NoError(t, err)
}
