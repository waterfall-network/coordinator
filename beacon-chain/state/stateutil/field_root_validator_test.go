package stateutil

import (
	"reflect"
	"strings"
	"testing"

	mathutil "github.com/waterfall-foundation/coordinator/math"
	ethpb "github.com/waterfall-foundation/coordinator/proto/prysm/v1alpha1"
	"github.com/waterfall-foundation/coordinator/testing/assert"
)

func TestValidatorConstants(t *testing.T) {
	v := &ethpb.Validator{}
	refV := reflect.ValueOf(v).Elem()
	numFields := refV.NumField()
	numOfValFields := 0

	for i := 0; i < numFields; i++ {
		if strings.Contains(refV.Type().Field(i).Name, "state") ||
			strings.Contains(refV.Type().Field(i).Name, "sizeCache") ||
			strings.Contains(refV.Type().Field(i).Name, "unknownFields") {
			continue
		}
		numOfValFields++
	}
	assert.Equal(t, validatorFieldRoots, numOfValFields)
	assert.Equal(t, uint64(validatorFieldRoots), mathutil.PowerOf2(validatorTreeDepth))

	_, err := ValidatorRegistryRoot([]*ethpb.Validator{v})
	assert.NoError(t, err)
}
