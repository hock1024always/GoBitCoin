package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"bitcoin-model-blockchain/core"
	"bitcoin-model-blockchain/network"
)

const (
	nodeID = "client"
)

func main() {
	// 命令行参数
	var (
		createWallet = flag.Bool("createwallet", false, "Create a new wallet")
		address      = flag.String("address", "", "Your wallet address")
		to           = flag.String("to", "", "Recipient address")
		amount       = flag.Int("amount", 0, "Amount to send")
		balance      = flag.Bool("balance", false, "Check balance")
		listWallets  = flag.Bool("listwallets", false, "List all wallets")
		interactive  = flag.Bool("interactive", false, "Run in interactive mode")
	)
	flag.Parse()

	// 创建钱包
	if *createWallet {
		createNewWallet()
		return
	}

	// 列出所有钱包
	if *listWallets {
		listAllWallets()
		return
	}

	// 交互模式
	if *interactive {
		runInteractiveMode()
		return
	}

	// 检查余额
	if *balance {
		if *address == "" {
			fmt.Println("Please specify -address")
			return
		}
		checkBalance(*address)
		return
	}

	// 发送交易
	if *to != "" && *amount > 0 {
		if *address == "" {
			fmt.Println("Please specify -address (your wallet address)")
			return
		}
		sendTransaction(*address, *to, *amount)
		return
	}

	// 显示帮助
	printHelp()
}

// createNewWallet 创建新钱包
func createNewWallet() {
	wallet, err := core.NewWallet()
	if err != nil {
		fmt.Printf("Failed to create wallet: %v\n", err)
		return
	}

	// 保存钱包
	if err := wallet.SaveToFile(nodeID); err != nil {
		fmt.Printf("Failed to save wallet: %v\n", err)
		return
	}

	fmt.Println("Wallet created successfully!")
	fmt.Printf("Address: %s\n", wallet.Address)
	fmt.Printf("Public Key: %s\n", hex.EncodeToString(wallet.PublicKey))
	fmt.Printf("Wallet saved to: wallets/wallet_%s.dat\n", nodeID)
}

// listAllWallets 列出所有钱包
func listAllWallets() {
	// 检查默认钱包
	if core.WalletExists(nodeID) {
		wallet, err := core.LoadWalletFromFile(nodeID)
		if err != nil {
			fmt.Printf("Failed to load wallet: %v\n", err)
			return
		}
		fmt.Printf("Wallet: %s\n", wallet.Address)
		fmt.Printf("  Public Key: %s\n", hex.EncodeToString(wallet.PublicKey))
	} else {
		fmt.Println("No wallet found. Use -createwallet to create one.")
	}
}

// checkBalance 查询余额
func checkBalance(address string) {
	// 加载区块链
	bc, err := loadBlockchain()
	if err != nil {
		fmt.Printf("Failed to load blockchain: %v\n", err)
		return
	}

	balance := bc.UTXOSet.GetBalance(address)
	fmt.Printf("Address: %s\n", address)
	fmt.Printf("Balance: %d coins\n", balance)

	// 显示UTXO详情
	utxos := bc.UTXOSet.FindUTXO(address)
	if len(utxos) > 0 {
		fmt.Println("\nUTXOs:")
		for i, utxo := range utxos {
			fmt.Printf("  [%d] Value: %d, PubKeyHash: %s...\n",
				i, utxo.Value, hex.EncodeToString(utxo.PubKeyHash)[:16])
		}
	}
}

// sendTransaction 发送交易
func sendTransaction(from, to string, amount int) {
	// 加载钱包
	wallet, err := core.LoadWalletFromFile(nodeID)
	if err != nil {
		fmt.Printf("Failed to load wallet: %v\n", err)
		return
	}

	// 验证地址
	if wallet.Address != from {
		fmt.Println("Error: Specified address doesn't match wallet")
		return
	}

	// 加载区块链
	bc, err := loadBlockchain()
	if err != nil {
		fmt.Printf("Failed to load blockchain: %v\n", err)
		return
	}

	// 检查余额
	balance := bc.UTXOSet.GetBalance(from)
	if balance < amount {
		fmt.Printf("Insufficient balance. Have: %d, Need: %d\n", balance, amount)
		return
	}

	// 创建交易
	tx, err := core.NewUTXOTransaction(
		from,
		to,
		amount,
		bc.UTXOSet.UTXO,
		wallet.GetPrivateKeyBytes(),
		wallet.PublicKey,
	)
	if err != nil {
		fmt.Printf("Failed to create transaction: %v\n", err)
		return
	}

	// 添加到内存池
	mempool := network.NewMempool(nodeID)
	if err := mempool.AddTransaction(tx); err != nil {
		fmt.Printf("Failed to add transaction to mempool: %v\n", err)
		return
	}

	// 广播交易（简化版）
	if err := mempool.BroadcastTransaction(tx); err != nil {
		fmt.Printf("Warning: Failed to broadcast transaction: %v\n", err)
	}

	fmt.Println("Transaction created and broadcasted successfully!")
	fmt.Printf("Transaction ID: %s\n", hex.EncodeToString(tx.ID))
	fmt.Printf("From: %s\n", from)
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Amount: %d\n", amount)
	fmt.Println("Waiting for miner to include it in a block...")
}

// loadBlockchain 加载区块链
func loadBlockchain() (*core.Blockchain, error) {
	// 客户端使用矿工的区块链数据
	minerNodeID := "miner"

	// 尝试加载矿工的区块链
	return core.LoadBlockchain(minerNodeID)
}

// runInteractiveMode 运行交互模式
func runInteractiveMode() {
	fmt.Println("========================================")
	fmt.Println("  Bitcoin Model Blockchain Client")
	fmt.Println("========================================")
	fmt.Println()

	// 检查钱包
	if !core.WalletExists(nodeID) {
		fmt.Println("No wallet found. Creating a new one...")
		createNewWallet()
		fmt.Println()
	}

	wallet, err := core.LoadWalletFromFile(nodeID)
	if err != nil {
		fmt.Printf("Failed to load wallet: %v\n", err)
		return
	}

	fmt.Printf("Your Address: %s\n", wallet.Address)

	// 加载区块链查看余额
	bc, err := loadBlockchain()
	if err == nil {
		balance := bc.UTXOSet.GetBalance(wallet.Address)
		fmt.Printf("Your Balance: %d coins\n", balance)
	}
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\nCommands:")
		fmt.Println("  1. send    - Send coins to another address")
		fmt.Println("  2. balance - Check your balance")
		fmt.Println("  3. newaddr - Create a new wallet address")
		fmt.Println("  4. list    - List all wallet addresses")
		fmt.Println("  5. help    - Show help")
		fmt.Println("  6. quit    - Exit")
		fmt.Print("\nEnter command: ")

		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch command {
		case "1", "send":
			fmt.Print("Enter recipient address: ")
			if !scanner.Scan() {
				continue
			}
			to := strings.TrimSpace(scanner.Text())

			fmt.Print("Enter amount: ")
			if !scanner.Scan() {
				continue
			}
			amountStr := strings.TrimSpace(scanner.Text())
			amount, err := strconv.Atoi(amountStr)
			if err != nil || amount <= 0 {
				fmt.Println("Invalid amount")
				continue
			}

			sendTransaction(wallet.Address, to, amount)

		case "2", "balance":
			checkBalance(wallet.Address)

		case "3", "newaddr":
			createNewWallet()
			// 重新加载钱包
			wallet, _ = core.LoadWalletFromFile(nodeID)

		case "4", "list":
			listAllWallets()

		case "5", "help":
			printInteractiveHelp()

		case "6", "quit", "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("Bitcoin Model Blockchain Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client -createwallet                    Create a new wallet")
	fmt.Println("  client -listwallets                     List all wallets")
	fmt.Println("  client -balance -address=<addr>         Check balance")
	fmt.Println("  client -address=<addr> -to=<addr> -amount=<n>  Send coins")
	fmt.Println("  client -interactive                     Run in interactive mode")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client -createwallet")
	fmt.Println("  client -balance -address=abc123...")
	fmt.Println("  client -address=abc123... -to=def456... -amount=10")
	fmt.Println("  client -interactive")
}

// printInteractiveHelp 打印交互模式帮助
func printInteractiveHelp() {
	fmt.Println("\nHelp:")
	fmt.Println("  send    - Send coins to another address")
	fmt.Println("            You will be prompted for recipient and amount")
	fmt.Println("  balance - Check your current balance")
	fmt.Println("  newaddr - Create a new wallet address")
	fmt.Println("  list    - List all your wallet addresses")
	fmt.Println("  help    - Show this help message")
	fmt.Println("  quit    - Exit the program")
}
