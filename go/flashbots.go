package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/metachris/flashbotsrpc"
)

const (
	flashbotsRPC   = "https://relay-sepolia.flashbots.net"
	presaleNFTAddr = "0x421E9AcaaB5a10EC2338BAc06A27c34F045a6395"
)

var (
	contractAbi abi.ABI
	sepoliaRPC  string
)

func init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 设置 Sepolia RPC URL
	infuraProjectID := os.Getenv("INFURA_PROJECT_ID")
	if infuraProjectID == "" {
		log.Fatal("INFURA_PROJECT_ID not found in .env file")
	}
	sepoliaRPC = fmt.Sprintf("https://sepolia.infura.io/v3/%s", infuraProjectID)

	// 初始化合约 ABI
	var err error
	contractAbi, err = abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"address","name":"initialAdmin","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"EnforcedPause","type":"error"},{"inputs":[],"name":"ExpectedPause","type":"error"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"OwnableInvalidOwner","type":"error"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"OwnableUnauthorizedAccount","type":"error"},{"inputs":[],"name":"PresaleNotActive","type":"error"},{"inputs":[],"name":"ReentrancyGuardReentrantCall","type":"error"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"inputs":[{"internalType":"address payable","name":"recipient","type":"address"}],"name":"destroy","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_amount","type":"uint256"}],"name":"doPresale","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_address","type":"address"}],"name":"getPresaleInfo","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"isPresaleActive","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"presaleInfo","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bool","name":"_isPresaleActive","type":"bool"}],"name":"startPresale","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`))

	if err != nil {
		log.Fatalf("初始化合约 ABI 失败: %v", err)
	}
}

// 检查预售状态
func checkPresaleState(client *ethclient.Client, contractAddr string) (bool, error) {
	data, err := contractAbi.Pack("isPresaleActive")
	if err != nil {
		return false, err
	}

	to := common.HexToAddress(contractAddr)
	msg := ethereum.CallMsg{
		To:   &to,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, err
	}

	var isActive bool
	err = contractAbi.UnpackIntoInterface(&isActive, "isPresaleActive", result)
	return isActive, err
}

func sendFlashbotsBundle(client *ethclient.Client, contractAddr string, signingKey *ecdsa.PrivateKey) error {
	// 创建 Flashbots RPC 客户端
	flashbotsClient := flashbotsrpc.New(flashbotsRPC)

	// 准备 doPresale 调用数据
	amount := big.NewInt(1)
	data, err := contractAbi.Pack("doPresale", amount)
	if err != nil {
		return err
	}

	// 获取当前区块
	blockNum, err := client.BlockNumber(context.Background())
	if err != nil {
		return err
	}

	// 获取基础 gas 价格并提高它
	baseGasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	gasPrice := new(big.Int).Mul(baseGasPrice, big.NewInt(10)) // 10倍基础价格
	minGasPrice := big.NewInt(5000000000)                      // 至少 5 Gwei
	if gasPrice.Cmp(minGasPrice) < 0 {
		gasPrice = minGasPrice
	}

	// 在多个连续区块中尝试发送 bundle
	for i := int64(1); i <= 10; i++ {
		targetBlock := blockNum + uint64(i)

		// 获取 nonce
		nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(signingKey.PublicKey))
		if err != nil {
			return err
		}

		// 构建交易
		tx := types.NewTransaction(
			nonce,
			common.HexToAddress(contractAddr),
			big.NewInt(0),
			100000,
			gasPrice,
			data,
		)

		// 签名交易
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(11155111)), signingKey)
		if err != nil {
			return err
		}

		// 获取 raw 交易数据
		rawTx, err := signedTx.MarshalBinary()
		if err != nil {
			return err
		}

		// 准备 bundle 请求
		now := uint64(time.Now().Unix())
		future := uint64(time.Now().Add(time.Minute).Unix())

		sendBundleArgs := flashbotsrpc.FlashbotsSendBundleRequest{
			Txs:          []string{fmt.Sprintf("0x%x", rawTx)},
			BlockNumber:  fmt.Sprintf("0x%x", targetBlock),
			MinTimestamp: &now,
			MaxTimestamp: &future,
		}

		// 发送 bundle
		bundleResponse, err := flashbotsClient.FlashbotsSendBundle(signingKey, sendBundleArgs)
		if err != nil {
			log.Printf("区块 %d 发送失败: %v", targetBlock, err)
			continue
		}

		fmt.Printf("Bundle 已发送到区块 %d, Response: %+v\n", targetBlock, bundleResponse)

		// 等待区块被挖出
		time.Sleep(12 * time.Second)

		// 检查交易是否被打包
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(targetBlock)))
		if err != nil {
			continue
		}

		for _, tx := range block.Transactions() {
			if tx.Hash().Hex() == signedTx.Hash().Hex() {
				fmt.Printf("交易成功被打包在区块 %d 中!\n", targetBlock)
				return nil
			}
		}

		fmt.Printf("交易未被打包在区块 %d 中\n", targetBlock)
	}

	return fmt.Errorf("bundle 在多个区块中都未被打包")
}

func main() {
	// 从 .env 文件读取私钥
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("PRIVATE_KEY not found in .env file")
	}
	// 如果私钥包含 "0x" 前缀，移除它
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// 生成 Flashbots 签名私钥
	flashbotsSigningKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("生成 Flashbots 签名私钥失败: %v", err)
	}
	privateKeyBytes := crypto.FromECDSA(flashbotsSigningKey)
	fmt.Printf("生成的 Flashbots 签名私钥: 0x%x\n", privateKeyBytes)

	// 解析交易账户私钥
	signingKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("解析私钥失败: %v", err)
	}

	// 打印账户地址（可选，用于确认）
	address := crypto.PubkeyToAddress(signingKey.PublicKey)
	fmt.Printf("使用账户地址: %s\n", address.Hex())

	// 连接以太坊客户端
	client, err := ethclient.Dial(sepoliaRPC)
	if err != nil {
		log.Fatal(err)
	}

	// 定期检查状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastState bool
	fmt.Println("开始监控预售状态...")

	for range ticker.C {
		state, err := checkPresaleState(client, presaleNFTAddr)
		if err != nil {
			log.Printf("检查状态失败: %v", err)
			continue
		}

		if !lastState && state {
			fmt.Println("预售已开启! 准备发送交易...")
			// 使用 signingKey 来签名交易，使用 flashbotsSigningKey 来发送 bundle
			err := sendFlashbotsBundle(client, presaleNFTAddr, signingKey) // 这里改用 signingKey
			if err != nil {
				log.Printf("发送交易失败: %v", err)
			}
		}

		lastState = state
	}
}
