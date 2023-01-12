package features

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
)

func TestDeprecatedFlags(t *testing.T) {
	for _, f := range deprecatedFlags {
		fv := reflect.ValueOf(f)
		field := reflect.Indirect(fv).FieldByName("Hidden")
		assert.Equal(t, false, !field.IsValid() || !field.Bool())
		assert.Equal(t, false, !strings.Contains(reflect.Indirect(fv).FieldByName("Usage").String(), "DEPRECATED. DO NOT USE."))
	}
}
