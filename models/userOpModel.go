package models

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PackedUserOperation 结构体
type PackedUserOperation struct {
	Sender             common.Address `bson:"sender"`
	Nonce              *big.Int       `bson:"nonce"`
	InitCode           string         `bson:"initCode"`
	CallData           string         `bson:"callData"`
	AccountGasLimits   string         `bson:"accountGasLimits"`
	PreVerificationGas *big.Int       `bson:"preVerificationGas"`
	GasFees            string         `bson:"gasFees"`
	PaymasterAndData   string         `bson:"paymasterAndData"`
	Signature          string         `bson:"signature"`
}
