# Go Waterfall
## Building the source
**We strongly recommend installing go version 1.21.11 or later**

Install Basel version 6.4.0

```shell
bazel build //beacon-chain:beacon-chain --config=release
bazel build //validator:validator --config=release
```

# Prysm: An Ethereum Consensus Implementation Written in Go

[![Build status](https://badge.buildkite.com/b555891daf3614bae4284dcf365b2340cefc0089839526f096.svg?branch=master)](https://buildkite.com/prysmatic-labs/prysm)
[![Go Report Card](https://goreportcard.com/badge/gitlab.waterfall.network/waterfall/protocol/coordinator)](https://goreportcard.com/report/gitlab.waterfall.network/waterfall/protocol/coordinator)
[![Consensus_Spec_Version 1.1.8](https://img.shields.io/badge/Consensus%20Spec%20Version-v1.1.8-blue.svg)](https://github.com/ethereum/consensus-specs/tree/v1.1.8)
[![Discord](https://user-images.githubusercontent.com/7288322/34471967-1df7808a-efbb-11e7-9088-ed0b04151291.png)](https://discord.gg/CTYGPUJ)

This is the core repository for Prysm, a [Golang](https://golang.org/) implementation of the [Ethereum Consensus](https://ethereum.org/en/eth2/) specification, developed by [Prysmatic Labs](https://prysmaticlabs.com). See the [Changelog](https://gitlab.waterfall.network/waterfall/protocol/coordinator/releases) for details of the latest releases and upcoming breaking changes.

### Getting Started

A detailed set of installation and usage instructions as well as breakdowns of each individual component are available in the [official documentation portal](https://docs.prylabs.network). If you still have questions, feel free to stop by our [Discord](https://discord.gg/CTYGPUJ).

### Staking on Mainnet

To participate in staking, you can join the [official eth2 launchpad](https://launchpad.ethereum.org). The launchpad is the only recommended way to become a validator on mainnet. You can explore validator rewards/penalties via Bitfly's block explorer: [beaconcha.in](https://beaconcha.in), and follow the latest blocks added to the chain on [beaconscan](https://beaconscan.com).


## Contributing
### Branches
Prysm maintains two permanent branches:

* [master](https://gitlab.waterfall.network/waterfall/protocol/coordinator/tree/master): This points to the latest stable release. It is ideal for most users.
* [develop](https://gitlab.waterfall.network/waterfall/protocol/coordinator/tree/develop): This is used for development, it contains the latest PRs. Developers should base their PRs on this branch.

### Guide
Want to get involved? Check out our [Contribution Guide](https://docs.prylabs.network/docs/contribute/contribution-guidelines/) to learn more!

## License

[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html)

## Legal Disclaimer

[Terms of Use](/TERMS_OF_SERVICE.md)
