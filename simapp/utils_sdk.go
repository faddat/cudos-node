package simapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AccountAddressPrefix = "cudos"
)

var (
	AccountPubKeyPrefix    = AccountAddressPrefix + "pub"
	ValidatorAddressPrefix = AccountAddressPrefix + "valoper"
	ValidatorPubKeyPrefix  = AccountAddressPrefix + "valoperpub"
	ConsNodeAddressPrefix  = AccountAddressPrefix + "valcons"
	ConsNodePubKeyPrefix   = AccountAddressPrefix + "valconspub"
)

// Currently not used anythere
func InitializeSdk() {
	initializeSdkConstants()
	initializeSdkConfig()
}

func initializeSdkConstants() {
	sdk.DefaultBondDenom = "cudos"
	sdk.DefaultPowerReduction = sdk.NewIntFromUint64(1000000000000000000)
}

func initializeSdkConfig() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)
	config.Seal()
}