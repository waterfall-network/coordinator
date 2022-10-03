package validator

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/config/params"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(ioutil.Discard)
	// Use minimal config to reduce test setup time.
	prevConfig := params.BeaconConfig().Copy()
	defer params.OverrideBeaconConfig(prevConfig)
	params.OverrideBeaconConfig(params.MinimalSpecConfig())

	m.Run()
}
