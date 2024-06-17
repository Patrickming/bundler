package controllers

// import (
// 	"bytes"
// 	"context"
// 	"crypto/ecdsa"
// 	"database/sql"
// 	"fmt"
// 	"math/big"
// 	"net/http"
// 	"os"

// 	"bundler/config"
// 	"bundler/models"

// 	"github.com/ethereum/go-ethereum"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"github.com/ethereum/go-ethereum/crypto"
// 	"github.com/ethereum/go-ethereum/ethclient"
// 	"github.com/gin-gonic/gin"
// )

// const (
// 	privateKey           = ""                                           // 在 .env 文件中设置
// 	simpleStorageAddress = "0xef42b628e838ae8d082b7773c46e52878064b94e" // 合约地址
// )

// // SimpleStorageController 是处理 SimpleStorage 合约交互的控制器
// type SimpleStorageController struct {
// 	DB     *sql.DB
// 	Client *ethclient.Client
// }

// // NewSimpleStorageController 创建一个新的 SimpleStorageController 实例
// func NewSimpleStorageController(db *sql.DB) (*SimpleStorageController, error) {
// 	config.LoadEnv() // 加载环境变量
// 	rpcURL := config.GetEnv("RPC_URL")

// 	// 连接以太坊客户端
// 	client, err := ethclient.Dial(rpcURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("Failed to connect to the Ethereum client: %w", err)
// 	}

// 	return &SimpleStorageController{
// 		DB:     db,
// 		Client: client,
// 	}, nil
// }

// // SetSimpleStorageValue 处理设置合约中的值的请求
// func (ctrl *SimpleStorageController) SetSimpleStorageValue(c *gin.Context) {
// 	var request struct {
// 		NewValue uint64 `json:"newValue"` // 新值
// 	}

// 	// 绑定 JSON 请求数据
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	privateKey := config.GetEnv("PRIVATE_KEY") // 从环境变量获取私钥
// 	txHash, err := ctrl.callSimpleStorageSet(ctrl.Client, privateKey, request.NewValue)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// 将新值插入数据库
// 	query := `INSERT INTO SimpleStorage (storedValue) VALUES (?)`
// 	_, err = ctrl.DB.Exec(query, request.NewValue)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"transactionHash": txHash})
// }

// // GetSimpleStorageValue 处理获取合约中存储值的请求
// func (ctrl *SimpleStorageController) GetSimpleStorageValue(c *gin.Context) {
// 	value, err := ctrl.callSimpleStorageGet(ctrl.Client)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// 从数据库中获取最近的存储值
// 	var storedValue models.SimpleStorage
// 	query := `SELECT * FROM SimpleStorage WHERE storedValue = ? ORDER BY ID DESC LIMIT 1`
// 	row := ctrl.DB.QueryRow(query, value)
// 	err = row.Scan(&storedValue.ID, &storedValue.StoredValue)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			c.JSON(http.StatusOK, gin.H{"value": value, "dbValue": "no record found"})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		}
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"value": value, "dbValue": storedValue.StoredValue})
// }

// // callSimpleStorageSet 调用合约的 set 方法设置新值
// func (ctrl *SimpleStorageController) callSimpleStorageSet(client *ethclient.Client, privateKey string, newValue uint64) (string, error) {
// 	privateKeyECDSA, err := crypto.HexToECDSA(privateKey) // 将私钥转换为 ECDSA 格式
// 	if err != nil {
// 		return "", err
// 	}

// 	publicKey := privateKeyECDSA.Public()
// 	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
// 	if !ok {
// 		return "", fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
// 	}

// 	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
// 	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
// 	if err != nil {
// 		return "", err
// 	}

// 	gasPrice, err := client.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		return "", err
// 	}

// 	abiData, err := os.ReadFile("./abi/SimpleStorage.json") // 读取合约 ABI
// 	if err != nil {
// 		return "", err
// 	}

// 	contractAbi, err := abi.JSON(bytes.NewReader(abiData)) // 解析 ABI
// 	if err != nil {
// 		return "", err
// 	}

// 	data, err := contractAbi.Pack("set", new(big.Int).SetUint64(newValue)) // 打包数据
// 	if err != nil {
// 		return "", err
// 	}

// 	value := big.NewInt(0)
// 	gasLimit := uint64(300000)
// 	toAddress := common.HexToAddress(simpleStorageAddress)

// 	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data) // 创建交易
// 	chainID, err := client.NetworkID(context.Background())
// 	if err != nil {
// 		return "", err
// 	}

// 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKeyECDSA) // 签名交易
// 	if err != nil {
// 		return "", err
// 	}

// 	err = client.SendTransaction(context.Background(), signedTx) // 发送交易
// 	if err != nil {
// 		return "", err
// 	}

// 	return signedTx.Hash().Hex(), nil
// }

// // callSimpleStorageGet 调用合约的 get 方法获取存储的值
// func (ctrl *SimpleStorageController) callSimpleStorageGet(client *ethclient.Client) (uint64, error) {
// 	abiData, err := os.ReadFile("./abi/SimpleStorage.json") // 读取合约 ABI
// 	if err != nil {
// 		return 0, err
// 	}

// 	contractAbi, err := abi.JSON(bytes.NewReader(abiData)) // 解析 ABI
// 	if err != nil {
// 		return 0, err
// 	}

// 	toAddress := common.HexToAddress(simpleStorageAddress)
// 	data, err := contractAbi.Pack("get") // 打包数据
// 	if err != nil {
// 		return 0, err
// 	}

// 	callMsg := ethereum.CallMsg{
// 		To:   &toAddress,
// 		Data: data,
// 	}

// 	result, err := client.CallContract(context.Background(), callMsg, nil) // 调用合约
// 	if err != nil {
// 		return 0, err
// 	}

// 	var value *big.Int
// 	err = contractAbi.UnpackIntoInterface(&value, "get", result) // 解包数据
// 	if err != nil {
// 		return 0, err
// 	}

// 	return value.Uint64(), nil
// }
