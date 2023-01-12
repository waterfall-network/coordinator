package cmd

import (
	"flag"
	"testing"

	"github.com/urfave/cli/v2"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/config/params"
	"gitlab.waterfall.network/waterfall/protocol/coordinator/testing/assert"
)

func TestOverrideConfig(t *testing.T) {
	cfg := &Flags{
		MinimalConfig: true,
	}
	reset := InitWithReset(cfg)
	defer reset()
	c := Get()
	assert.Equal(t, true, c.MinimalConfig)
}

func TestDefaultConfig(t *testing.T) {
	cfg := &Flags{
		MaxRPCPageSize: params.BeaconConfig().DefaultPageSize,
	}
	c := Get()
	assert.DeepEqual(t, c, cfg)

	reset := InitWithReset(cfg)
	defer reset()
	c = Get()
	assert.DeepEqual(t, c, cfg)
}

func TestConfigureBeaconConfig(t *testing.T) {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.Bool(MinimalConfigFlag.Name, true, "test")
	context := cli.NewContext(&app, set, nil)
	ConfigureBeaconChain(context)
	c := Get()
	assert.Equal(t, true, c.MinimalConfig)
}
