package deposit

import (
	"gitlab.waterfall.network/waterfall/protocol/gwat/accounts/abi/bind"
)

// NewDepositContractcallFromBoundContract creates a new instance of DepositContractCaller, bound to
// a specific deployed contract.
func NewDepositContractCallerFromBoundContract(contract *bind.BoundContract) DepositContractCaller {
	return DepositContractCaller{contract: contract}
}

// NewDepositContractTransactorFromBoundContract creates a new instance of
// DepositContractTransactor, bound to a specific deployed contract.
func NewDepositContractTransactorFromBoundContract(contract *bind.BoundContract) DepositContractTransactor {
	return DepositContractTransactor{contract: contract}
}

// NewDepositContractFiltererFromBoundContract creates a new instance of
// DepositContractFilterer, bound to a specific deployed contract.
func NewDepositContractFiltererFromBoundContract(contract *bind.BoundContract) DepositContractFilterer {
	return DepositContractFilterer{contract: contract}
}
