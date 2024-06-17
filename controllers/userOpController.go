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
	"time"

	"bundler/config"
	"bundler/models"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	entryPointAddress = "0xF988D980A36c3E8da79AB91B4562fD81adA7ECE3" // EntryPoint 合约地址
)

type UserOpController struct {
	Client     *ethclient.Client
	Collection *mongo.Collection
}

// NewUserOpController 创建一个新的 UserOpController 实例
func NewUserOpController() (*UserOpController, error) {
	rpcURL := os.Getenv("RPC_URL") // 从环境变量中读取 RPC URL

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the Ethereum client: %w", err)
	}

	mongoClient, err := config.GetMongoClient()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to MongoDB: %w", err)
	}

	collection := mongoClient.Database("userop_db").Collection("userops")

	return &UserOpController{
		Client:     client,
		Collection: collection,
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

	// 验证并解码每个字段的十六进制字符串
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

	accountGasLimits, err := validateAndDecodeHexString(userOp.AccountGasLimits)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid accountGasLimits: %v", err)})
		return
	}

	gasFees, err := validateAndDecodeHexString(userOp.GasFees)
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

	// 确保存储和输出数据的一致性
	userOp.Nonce = big.NewInt(userOp.Nonce.Int64())

	// 存储到 MongoDB
	userOp.InitCode = hex.EncodeToString(initCode)
	userOp.CallData = hex.EncodeToString(callData)
	userOp.AccountGasLimits = hex.EncodeToString(accountGasLimits)
	userOp.GasFees = hex.EncodeToString(gasFees)
	userOp.PaymasterAndData = hex.EncodeToString(paymasterAndData)
	userOp.Signature = hex.EncodeToString(signature)

	_, err = ctrl.Collection.InsertOne(context.TODO(), userOp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to store UserOp: %v", err)})
		return
	}

	// 处理并发送 UserOp
	time.Sleep(1 * time.Second) // 模拟处理时间
	txHash, err := ctrl.processAndSendUserOp(userOp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 删除已处理的 userOp
	filter := bson.M{"nonce": userOp.Nonce}
	_, err = ctrl.Collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete UserOp: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "UserOp received and sent", "transactionHash": txHash})
}

// validateAndDecodeHexString 验证并解码十六进制字符串
func validateAndDecodeHexString(hexString string) ([]byte, error) {
	if len(hexString) < 2 || hexString[:2] != "0x" {
		return nil, fmt.Errorf("invalid hex string: %s", hexString)
	}
	decoded, err := hex.DecodeString(hexString[2:])
	if err != nil {
		return nil, fmt.Errorf("error decoding hex string: %v", err)
	}
	return decoded, nil
}

// processAndSendUserOp 处理并发送 UserOp 到区块链
func (ctrl *UserOpController) processAndSendUserOp(userOp models.PackedUserOperation) (string, error) {
	privateKey := os.Getenv("PRIVATE_KEY") // 从环境变量中读取私钥
	abiPath := "./abi/EntryPoint.json"     // 合约 ABI 文件路径

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

	// 将存储在 MongoDB 中的字符串字段转换回字节切片
	initCodeBytes, err := hex.DecodeString(userOp.InitCode)
	if err != nil {
		return "", fmt.Errorf("error decoding initCode: %v", err)
	}

	callDataBytes, err := hex.DecodeString(userOp.CallData)
	if err != nil {
		return "", fmt.Errorf("error decoding callData: %v", err)
	}

	accountGasLimitsBytes, err := hex.DecodeString(userOp.AccountGasLimits)
	if err != nil {
		return "", fmt.Errorf("error decoding accountGasLimits: %v", err)
	}

	gasFeesBytes, err := hex.DecodeString(userOp.GasFees)
	if err != nil {
		return "", fmt.Errorf("error decoding gasFees: %v", err)
	}

	paymasterAndDataBytes, err := hex.DecodeString(userOp.PaymasterAndData)
	if err != nil {
		return "", fmt.Errorf("error decoding paymasterAndData: %v", err)
	}

	signatureBytes, err := hex.DecodeString(userOp.Signature)
	if err != nil {
		return "", fmt.Errorf("error decoding signature: %v", err)
	}

	// 输出用于调试的实际发送数据
	fmt.Printf("Prepared UserOp for contract call:\n{\n\tSender: \"%s\",\n\tNonce: %s,\n\tInitCode: \"%s\",\n\tCallData: \"%s\",\n\tAccountGasLimits: \"%s\",\n\tPreVerificationGas: %d,\n\tGasFees: \"%s\",\n\tPaymasterAndData: \"%s\",\n\tSignature: \"%s\"\n}\n",
		userOp.Sender.Hex(),
		userOp.Nonce.String(),
		userOp.InitCode,
		userOp.CallData,
		userOp.AccountGasLimits,
		userOp.PreVerificationGas.Uint64(),
		userOp.GasFees,
		userOp.PaymasterAndData,
		userOp.Signature)

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
	}{
		{
			Sender:             userOp.Sender,
			Nonce:              userOp.Nonce,
			InitCode:           initCodeBytes,
			CallData:           callDataBytes,
			AccountGasLimits:   toFixedSizeByteArray(accountGasLimitsBytes),
			PreVerificationGas: userOp.PreVerificationGas,
			GasFees:            toFixedSizeByteArray(gasFeesBytes),
			PaymasterAndData:   paymasterAndDataBytes,
			Signature:          signatureBytes,
		},
	}, common.HexToAddress("0xC26Cbf92EdD4D0bE0d73264f097F76432ffb81D1"))
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
	fmt.Printf("Transaction sent with hash: %s\n", signedTx.Hash().Hex())
	return signedTx.Hash().Hex(), nil
}

// toFixedSizeByteArray 将字节切片转换为固定大小的字节数组
func toFixedSizeByteArray(data []byte) [32]byte {
	var array [32]byte
	copy(array[:], data)
	return array
}
