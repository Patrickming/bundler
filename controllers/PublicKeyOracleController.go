// PublicKeyOracleController.go

package controllers

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	publicKeyOracleAddress = "0x57A9Edbb9fF61EFFB33537994f7F0E1fabBaA282" // 更新为正确的 PublicKeyOracle 合约地址
)

type PublicKeyOracleController struct {
	Client *ethclient.Client
}

// NewPublicKeyOracleController 创建一个新的 PublicKeyOracleController 实例
func NewPublicKeyOracleController() (*PublicKeyOracleController, error) {
	rpcURL := os.Getenv("RPC_URL") // 从环境变量中读取 RPC URL

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the Ethereum client: %w", err)
	}
	return &PublicKeyOracleController{
		Client: client,
	}, nil
}

// SetPublicKey 调用合约的 setPublicKey 方法，设置公钥信息
func (ctrl *PublicKeyOracleController) SetPublicKey(domain, selector string, modulus, exponent []byte) (string, error) {
	privateKey := os.Getenv("PRIVATE_KEY")
	abiPath := os.Getenv("PublicKeyOracle_ABI") // 合约 ABI 文件路径

	// 将私钥字符串转换为 ECDSA 私钥
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("error converting private key: %v", err)
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("error asserting type of public key")
	}

	// 从公钥推导出以太坊地址
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 读取并解析 ABI 文件
	abiData, err := os.ReadFile(abiPath)
	if err != nil {
		return "", fmt.Errorf("error reading ABI file: %v", err)
	}

	var contractAbi abi.ABI
	if err := contractAbi.UnmarshalJSON(abiData); err != nil {
		return "", fmt.Errorf("error parsing ABI: %v", err)
	}

	// 使用 ABI 打包数据以调用 setPublicKey 方法
	data, err := contractAbi.Pack("setPublicKey", domain, selector, modulus, exponent)
	if err != nil {
		return "", fmt.Errorf("error packing data: %v", err)
	}

	// 获取账户的当前 nonce
	nonce, err := ctrl.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("error getting nonce: %v", err)
	}

	// 获取建议的 gas price
	gasPrice, err := ctrl.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("error getting gas price: %v", err)
	}

	// 计算发送交易需要的 gas limit
	gasLimit := uint64(200000) // 根据实际情况调整

	// 创建交易对象
	value := big.NewInt(0)
	toAddress := common.HexToAddress(publicKeyOracleAddress)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	// 获取区块链的 chain ID
	chainID, err := ctrl.Client.NetworkID(context.Background())
	if err != nil {
		return "", fmt.Errorf("error getting network ID: %v", err)
	}

	// 签署交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKeyECDSA)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	// 发送交易到区块链
	err = ctrl.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("error sending transaction: %v", err)
	}

	// 输出交易哈希
	fmt.Printf("SetPublicKey transaction sent with hash: %s\n", signedTx.Hash().Hex())
	return signedTx.Hash().Hex(), nil
}

// GetRSAKey 方法调用合约的 getRSAKey 方法，获取公钥信息
func (ctrl *PublicKeyOracleController) GetRSAKey(domain, selector string) (string, []byte, []byte, error) {
	// 读取合约 ABI 文件路径
	abiPath := os.Getenv("PublicKeyOracle_ABI")
	abiData, err := os.ReadFile(abiPath)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error reading ABI file: %v", err)
	}

	var contractAbi abi.ABI
	if err := json.Unmarshal(abiData, &contractAbi); err != nil {
		return "", nil, nil, fmt.Errorf("error parsing ABI: %v", err)
	}

	// 获取当前账户的 nonce
	privateKey := os.Getenv("PRIVATE_KEY")
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error converting private key: %v", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)
	nonce, err := ctrl.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error getting nonce: %v", err)
	}

	// 获取建议的 gas price
	gasPrice, err := ctrl.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", nil, nil, fmt.Errorf("error getting gas price: %v", err)
	}

	// 计算发送交易需要的 gas limit
	gasLimit := uint64(200000) // 根据实际情况调整

	// 构造调用数据
	callData, err := contractAbi.Pack("getRSAKey", domain, selector)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error packing data: %v", err)
	}

	// 创建交易对象
	value := big.NewInt(0)
	toAddress := common.HexToAddress(publicKeyOracleAddress)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, callData)

	// 获取网络 ID
	networkId, err := ctrl.Client.NetworkID(context.Background())
	if err != nil {
		return "", nil, nil, fmt.Errorf("error getting network ID: %v", err)
	}

	// 签署交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(networkId), privateKeyECDSA)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error signing transaction: %v", err)
	}

	// 发送交易到区块链
	err = ctrl.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error sending transaction: %v", err)
	}

	// 解析交易事件数据  这里是如果没有返回数据就会报错，所以先注释
	// 等待交易完成
	// receipt, err := bind.WaitMined(context.Background(), ctrl.Client, signedTx)
	// if err != nil {
	// 	return "", nil, nil, fmt.Errorf("error waiting for transaction to be mined: %v", err)
	// }

	// 解析交易事件数据  这里是如果没有返回数据就会报错，所以先注释
	// var result map[string]interface{}
	//如果没有接收到数据就直接报错（后加）
	// if len(receipt.Logs) == 0 {
	// 	return "", nil, nil, fmt.Errorf("no logs found in receipt")
	// }

	// err = contractAbi.UnpackIntoMap(result, "getRSAKey", receipt.Logs[0].Data)
	// if err != nil {
	// 	return "", nil, nil, fmt.Errorf("error unpacking result: %v", err)
	// }

	// // 提取结果中的 Modulus 和 Exponent
	// modulus, ok := result["Modulus"].([]byte)
	// if !ok {
	// 	return "", nil, nil, fmt.Errorf("error casting modulus to []byte")
	// }

	// exponent, ok := result["Exponent"].([]byte)
	// if !ok {
	// 	return "", nil, nil, fmt.Errorf("error casting exponent to []byte")
	// }

	// 返回交易哈希和结果
	txHash := signedTx.Hash().Hex()

	// return txHash, modulus, exponent, nil
	return txHash, nil, nil, nil //无数据暂时返回空

}
