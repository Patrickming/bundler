package models

// type PackedUserOperation struct {
// 	Sender             string   `json:"sender"`
// 	Nonce              uint64   `json:"nonce"`
// 	InitCode           []byte   `json:"initCode"`
// 	CallData           []byte   `json:"callData"`
// 	AccountGasLimits   [32]byte `json:"accountGasLimits"`
// 	PreVerificationGas uint64   `json:"preVerificationGas"`
// 	GasFees            [32]byte `json:"gasFees"`
// 	PaymasterAndData   []byte   `json:"paymasterAndData"`
// 	Signature          []byte   `json:"signature"`
// }

// type PackedUserOperation struct {
// 	Sender             string `json:"sender"`
// 	Nonce              uint64 `json:"nonce"`
// 	InitCode           string `json:"initCode"`         // 临时字符串类型，用于接收 JSON 数据
// 	CallData           string `json:"callData"`         // 临时字符串类型，用于接收 JSON 数据
// 	AccountGasLimits   string `json:"accountGasLimits"` // 临时字符串类型，用于接收 JSON 数据
// 	PreVerificationGas uint64 `json:"preVerificationGas"`
// 	GasFees            string `json:"gasFees"`          // 临时字符串类型，用于接收 JSON 数据
// 	PaymasterAndData   string `json:"paymasterAndData"` // 临时字符串类型，用于接收 JSON 数据
// 	Signature          string `json:"signature"`        // 临时字符串类型，用于接收 JSON 数据
// }

type PackedUserOperation struct {
	Sender             string `json:"sender"`
	Nonce              uint64 `json:"nonce"`
	InitCode           string `json:"initCode"`
	CallData           string `json:"callData"`
	AccountGasLimits   string `json:"accountGasLimits"`
	PreVerificationGas uint64 `json:"preVerificationGas"`
	GasFees            string `json:"gasFees"`
	PaymasterAndData   string `json:"paymasterAndData"`
	Signature          string `json:"signature"`
}
