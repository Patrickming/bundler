package controllers

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"bundler/config"
	"bundler/models"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

const (
	entryPointAddress = "0xF988D980A36c3E8da79AB91B4562fD81adA7ECE3" // EntryPoint 合约地址
)

type UserOpController struct {
	Client *ethclient.Client
}

// NewUserOpController 创建一个新的 UserOpController 实例
func NewUserOpController() (*UserOpController, error) {
	config.LoadEnv()
	rpcURL := config.GetEnv("RPC_URL") // 从环境变量中读取 RPC URL

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the Ethereum client: %w", err)
	}

	return &UserOpController{
		Client: client,
	}, nil
}

// StoreUserOp 处理接收到的 UserOp 请求
func (ctrl *UserOpController) StoreUserOp(c *gin.Context) {
	var userOp models.PackedUserOperation

	// 绑定 JSON 请求体到 userOp 结构体
	if err := c.ShouldBindJSON(&userOp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 输出接收到的 UserOp 用于调试
	fmt.Println("Received UserOp:", userOp)

	// 验证和解码每个字段的十六进制字符串
	initCode, err := validateAndDecodeHexString(userOp.InitCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid initCode: %v", err)})
		return
	}

	callData, err := validateAndDecodeHexString(userOp.CallData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid callData: %v", err)})
		return
	}

	accountGasLimits, err := validateAndDecodeFixedSizeHexString(userOp.AccountGasLimits, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid accountGasLimits: %v", err)})
		return
	}

	gasFees, err := validateAndDecodeFixedSizeHexString(userOp.GasFees, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid gasFees: %v", err)})
		return
	}

	paymasterAndData, err := validateAndDecodeHexString(userOp.PaymasterAndData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid paymasterAndData: %v", err)})
		return
	}

	signature, err := validateAndDecodeHexString(userOp.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid signature: %v", err)})
		return
	}

	// 处理并发送 UserOp
	txHash, err := ctrl.processAndSendUserOp(userOp.Sender, userOp.Nonce, initCode, callData, accountGasLimits, userOp.PreVerificationGas, gasFees, paymasterAndData, signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "UserOp received and sent", "transactionHash": txHash})
}

// validateAndDecodeHexString 验证并解码十六进制字符串
func validateAndDecodeHexString(hexString string) ([]byte, error) {
	if len(hexString) < 2 || hexString[:2] != "0x" {
		fmt.Println("Invalid hex string format:", hexString)
		return nil, fmt.Errorf("invalid hex string")
	}
	decoded, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Println("Error decoding hex string:", err)
		return nil, err
	}
	fmt.Println("Decoded hex string:", decoded)
	return decoded, nil
}

// validateAndDecodeFixedSizeHexString 验证并解码固定大小的十六进制字符串
func validateAndDecodeFixedSizeHexString(hexString string, size int) ([32]byte, error) {
	var fixedArray [32]byte
	decoded, err := validateAndDecodeHexString(hexString)
	if err != nil {
		return fixedArray, err
	}
	if len(decoded) != size {
		fmt.Printf("Invalid hex string size: got %d, expected %d\n", len(decoded), size)
		return fixedArray, fmt.Errorf("invalid hex string size")
	}
	copy(fixedArray[:], decoded)
	fmt.Println("Decoded fixed size hex string:", fixedArray)
	return fixedArray, nil
}

// processAndSendUserOp 处理并发送 UserOp 到区块链
func (ctrl *UserOpController) processAndSendUserOp(sender string, nonce uint64, initCode, callData []byte, accountGasLimits [32]byte, preVerificationGas uint64, gasFees [32]byte, paymasterAndData, signature []byte) (string, error) {
	privateKey := config.GetEnv("PRIVATE_KEY") // 从环境变量中读取私钥
	abiPath := "./abi/EntryPoint.json"         // 合约 ABI 文件路径

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
	if err := json.Unmarshal(abiData, &contractAbi); err != nil {
		return "", fmt.Errorf("error parsing ABI: %v", err)
	}

	// 创建 userOp 结构体
	userOp := struct {
		Sender             common.Address
		Nonce              *big.Int
		InitCode           []byte
		CallData           []byte
		AccountGasLimits   [32]byte
		PreVerificationGas *big.Int
		GasFees            [32]byte
		PaymasterAndData   []byte
		Signature          []byte
	}{
		Sender:             common.HexToAddress(sender),
		Nonce:              big.NewInt(int64(nonce)),
		InitCode:           initCode,
		CallData:           callData,
		AccountGasLimits:   accountGasLimits,
		PreVerificationGas: big.NewInt(int64(preVerificationGas)),
		GasFees:            gasFees,
		PaymasterAndData:   paymasterAndData,
		Signature:          signature,
	}

	// 使用 ABI 打包数据以调用 handleOps 方法
	data, err := contractAbi.Pack("handleOps", []struct {
		Sender             common.Address
		Nonce              *big.Int
		InitCode           []byte
		CallData           []byte
		AccountGasLimits   [32]byte
		PreVerificationGas *big.Int
		GasFees            [32]byte
		PaymasterAndData   []byte
		Signature          []byte
	}{userOp}, common.HexToAddress("0xC26Cbf92EdD4D0bE0d73264f097F76432ffb81D1"))
	if err != nil {
		return "", fmt.Errorf("error packing data: %v", err)
	}

	// 获取账户的当前 nonce
	nonce, err = ctrl.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("error getting nonce: %v", err)
	}

	// 获取建议的 gas price
	gasPrice, err := ctrl.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("error getting gas price: %v", err)
	}

	// 创建交易对象
	value := big.NewInt(0)
	gasLimit := uint64(300000)
	toAddress := common.HexToAddress(entryPointAddress)
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
	fmt.Println("Transaction sent with hash:", signedTx.Hash().Hex())
	return signedTx.Hash().Hex(), nil
}

/**
配置与初始化：代码通过 config.LoadEnv 函数加载环境变量，其中包括 RPC_URL 和 PRIVATE_KEY。
NewUserOpController 函数用来创建一个新的控制器实例，并连接到以太坊客户端。

处理请求：StoreUserOp 函数负责处理 HTTP POST 请求，解析并验证 JSON 请求体，将其转换为适当的数据格式，
并调用 processAndSendUserOp 函数处理 User Operation。

验证与解码：validateAndDecodeHexString 和 validateAndDecodeFixedSizeHexString 函数用于验证并解码十六进制字符串，
确保其格式正确并转换为字节数组。

交易处理：processAndSendUserOp 函数创建并签署以太坊交易，将 User Operation 发送到指定的 EntryPoint 合约地址。

注释与调试：代码在关键步骤中添加了详细的注释和调试信息，帮助理解和排查可能出现的问题。
*/
