package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"bitcoin-model-blockchain/core"
	"bitcoin-model-blockchain/network"
)

const (
	nodeID = "miner"
)

func main() {
	// 命令行参数
	var (
		createWallet = flag.Bool("createwallet", false, "Create a new wallet for miner")
		initChain    = flag.Bool("init", false, "Initialize blockchain with genesis block")
		mine         = flag.Bool("mine", false, "Start mining")
		showChain    = flag.Bool("showchain", false, "Show blockchain")
		showMempool  = flag.Bool("showmempool", false, "Show mempool")
		showBalance  = flag.Bool("balance", false, "Show miner balance")
		interactive  = flag.Bool("interactive", false, "Run in interactive mode")
		numBlocks    = flag.Int("n", 1, "Number of blocks to mine")
	)
	flag.Parse()

	// 创建矿工钱包
	if *createWallet {
		createMinerWallet()
		return
	}

	// 初始化区块链
	if *initChain {
		initializeBlockchain()
		return
	}

	// 显示区块链
	if *showChain {
		showBlockchain()
		return
	}

	// 显示内存池
	if *showMempool {
		showMempoolContent()
		return
	}

	// 显示余额
	if *showBalance {
		showMinerBalance()
		return
	}

	// 挖矿
	if *mine {
		startMining(*numBlocks)
		return
	}

	// 交互模式
	if *interactive {
		runInteractiveMode()
		return
	}

	// 显示帮助
	printHelp()
}

// createMinerWallet 创建矿工钱包
func createMinerWallet() {
	wallet, err := core.NewWallet()
	if err != nil {
		fmt.Printf("Failed to create wallet: %v\n", err)
		return
	}

	if err := wallet.SaveToFile(nodeID); err != nil {
		fmt.Printf("Failed to save wallet: %v\n", err)
		return
	}

	fmt.Println("Miner wallet created successfully!")
	fmt.Printf("Address: %s\n", wallet.Address)
	fmt.Printf("Public Key: %s\n", hex.EncodeToString(wallet.PublicKey))
}

// initializeBlockchain 初始化区块链
func initializeBlockchain() {
	// 检查是否已有钱包
	if !core.WalletExists(nodeID) {
		fmt.Println("No miner wallet found. Creating one...")
		createMinerWallet()
	}

	wallet, err := core.LoadWalletFromFile(nodeID)
	if err != nil {
		fmt.Printf("Failed to load wallet: %v\n", err)
		return
	}

	// 创建区块链（创世区块）
	bc, err := core.NewBlockchain(wallet.Address, nodeID)
	if err != nil {
		fmt.Printf("Failed to create blockchain: %v\n", err)
		return
	}

	fmt.Println("Blockchain initialized successfully!")
	fmt.Printf("Genesis block created with miner address: %s\n", wallet.Address)
	fmt.Printf("Block Hash: %s\n", hex.EncodeToString(bc.GetLatestBlock().Hash))
}

// showBlockchain 显示区块链
func showBlockchain() {
	bc, err := core.LoadBlockchain(nodeID)
	if err != nil {
		fmt.Printf("Failed to load blockchain: %v\n", err)
		return
	}

	fmt.Println("========================================")
	fmt.Println("           BLOCKCHAIN")
	fmt.Println("========================================")
	fmt.Printf("Height: %d\n", bc.GetHeight())
	fmt.Printf("Total Blocks: %d\n", len(bc.Blocks))
	fmt.Println()

	for _, block := range bc.Blocks {
		printBlock(block)
		fmt.Println("----------------------------------------")
	}
}

// printBlock 打印区块信息
func printBlock(block *core.Block) {
	fmt.Printf("Block #%d\n", block.Height)
	fmt.Printf("  Hash:        %s\n", hex.EncodeToString(block.Hash))
	fmt.Printf("  Prev Hash:   %s\n", hex.EncodeToString(block.Header.PrevBlockHash))
	fmt.Printf("  Merkle Root: %s\n", hex.EncodeToString(block.Header.MerkleRoot))
	fmt.Printf("  Timestamp:   %d\n", block.Header.Timestamp)
	fmt.Printf("  Difficulty:  %d\n", block.Header.Difficulty)
	fmt.Printf("  Nonce:       %d\n", block.Header.Nonce)
	fmt.Printf("  Transactions: %d\n", len(block.Transactions))

	for i, tx := range block.Transactions {
		fmt.Printf("\n  Transaction #%d:\n", i)
		fmt.Printf("    ID: %s\n", hex.EncodeToString(tx.ID))
		if tx.IsCoinbase() {
			fmt.Printf("    Type: Coinbase (Mining Reward)\n")
		} else {
			fmt.Printf("    Type: Regular\n")
		}
		fmt.Printf("    Inputs: %d\n", len(tx.Inputs))
		for j, in := range tx.Inputs {
			txIDStr := hex.EncodeToString(in.TxID)
			if len(txIDStr) > 16 {
				txIDStr = txIDStr[:16]
			}
			fmt.Printf("      [%d] TxID: %s..., Vout: %d\n",
				j, txIDStr, in.Vout)
		}
		fmt.Printf("    Outputs: %d\n", len(tx.Outputs))
		for j, out := range tx.Outputs {
			fmt.Printf("      [%d] Value: %d, PubKeyHash: %s...\n",
				j, out.Value, hex.EncodeToString(out.PubKeyHash)[:16])
		}
	}
}

// showMempoolContent 显示内存池内容
func showMempoolContent() {
	mempool := network.NewMempool(nodeID)

	// 接收广播的交易
	mempool.ReceiveBroadcastedTransactions()

	fmt.Println("========================================")
	fmt.Println("           MEMPOOL")
	fmt.Println("========================================")
	fmt.Printf("Pending Transactions: %d\n", mempool.Count())
	fmt.Println()

	if mempool.Count() == 0 {
		fmt.Println("No pending transactions")
		return
	}

	transactions := mempool.GetAllTransactions()
	for i, tx := range transactions {
		fmt.Printf("Transaction #%d:\n", i+1)
		fmt.Printf("  ID: %s\n", hex.EncodeToString(tx.ID))
		fmt.Printf("  Inputs: %d\n", len(tx.Inputs))
		fmt.Printf("  Outputs: %d\n", len(tx.Outputs))
		totalOutput := 0
		for _, out := range tx.Outputs {
			totalOutput += out.Value
		}
		fmt.Printf("  Total Output: %d\n", totalOutput)
		fmt.Println()
	}
}

// showMinerBalance 显示矿工余额
func showMinerBalance() {
	if !core.WalletExists(nodeID) {
		fmt.Println("No miner wallet found")
		return
	}

	wallet, err := core.LoadWalletFromFile(nodeID)
	if err != nil {
		fmt.Printf("Failed to load wallet: %v\n", err)
		return
	}

	bc, err := core.LoadBlockchain(nodeID)
	if err != nil {
		fmt.Printf("Failed to load blockchain: %v\n", err)
		return
	}

	balance := bc.UTXOSet.GetBalance(wallet.Address)

	fmt.Println("========================================")
	fmt.Println("           MINER BALANCE")
	fmt.Println("========================================")
	fmt.Printf("Address: %s\n", wallet.Address)
	fmt.Printf("Balance: %d coins\n", balance)

	// 显示UTXO详情
	utxos := bc.UTXOSet.FindUTXO(wallet.Address)
	if len(utxos) > 0 {
		fmt.Println("\nUTXOs:")
		for i, utxo := range utxos {
			fmt.Printf("  [%d] Value: %d coins\n", i, utxo.Value)
		}
	}
}

// startMining 开始挖矿
func startMining(numBlocks int) {
	// 检查钱包
	if !core.WalletExists(nodeID) {
		fmt.Println("No miner wallet found. Creating one...")
		createMinerWallet()
	}

	wallet, err := core.LoadWalletFromFile(nodeID)
	if err != nil {
		fmt.Printf("Failed to load wallet: %v\n", err)
		return
	}

	// 加载区块链
	bc, err := core.LoadBlockchain(nodeID)
	if err != nil {
		fmt.Printf("Failed to load blockchain: %v\n", err)
		return
	}

	// 加载内存池
	mempool := network.NewMempool(nodeID)

	// 创建矿工
	miner := core.NewMiner(wallet.Address, bc, mempool, bc.UTXOSet)

	fmt.Println("========================================")
	fmt.Println("           MINING STARTED")
	fmt.Println("========================================")
	fmt.Printf("Miner Address: %s\n", wallet.Address)
	fmt.Printf("Current Balance: %d coins\n", bc.UTXOSet.GetBalance(wallet.Address))
	fmt.Printf("Blockchain Height: %d\n", bc.GetHeight())
	fmt.Printf("Pending Transactions: %d\n", mempool.Count())
	fmt.Printf("Target Blocks: %d\n", numBlocks)
	fmt.Println()

	for i := 0; i < numBlocks; i++ {
		// 接收广播的交易
		mempool.ReceiveBroadcastedTransactions()

		if mempool.Count() == 0 {
			fmt.Printf("\nNo transactions to mine. Waiting...\n")
			fmt.Println("You can use the client to create transactions.")
			break
		}

		fmt.Printf("\n--- Mining Block #%d ---\n", bc.GetHeight()+1)
		startTime := time.Now()

		block, err := miner.MineBlock()
		if err != nil {
			fmt.Printf("Mining failed: %v\n", err)
			continue
		}

		duration := time.Since(startTime)
		fmt.Printf("Block mined in %v!\n", duration)
		fmt.Printf("Block Hash: %s\n", hex.EncodeToString(block.Hash))
		fmt.Printf("Transactions: %d\n", len(block.Transactions))
		fmt.Printf("New Balance: %d coins\n", bc.UTXOSet.GetBalance(wallet.Address))
	}

	fmt.Println("\n========================================")
	fmt.Println("           MINING COMPLETED")
	fmt.Println("========================================")
	fmt.Printf("Final Balance: %d coins\n", bc.UTXOSet.GetBalance(wallet.Address))
	fmt.Printf("Final Height: %d\n", bc.GetHeight())
}

// runInteractiveMode 运行交互模式
func runInteractiveMode() {
	fmt.Println("========================================")
	fmt.Println("  Bitcoin Model Blockchain Miner")
	fmt.Println("========================================")
	fmt.Println()

	// 检查钱包
	if !core.WalletExists(nodeID) {
		fmt.Println("No miner wallet found. Creating one...")
		createMinerWallet()
		fmt.Println()
	}

	wallet, err := core.LoadWalletFromFile(nodeID)
	if err != nil {
		fmt.Printf("Failed to load wallet: %v\n", err)
		return
	}

	fmt.Printf("Miner Address: %s\n", wallet.Address)

	// 检查区块链
	bc, err := core.LoadBlockchain(nodeID)
	if err != nil {
		fmt.Println("Blockchain not found. Initializing...")
		initializeBlockchain()
		bc, _ = core.LoadBlockchain(nodeID)
	}

	fmt.Printf("Blockchain Height: %d\n", bc.GetHeight())
	fmt.Printf("Current Balance: %d coins\n", bc.UTXOSet.GetBalance(wallet.Address))
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\nCommands:")
		fmt.Println("  1. mine     - Mine a new block")
		fmt.Println("  2. chain    - Show blockchain")
		fmt.Println("  3. mempool  - Show pending transactions")
		fmt.Println("  4. balance  - Show miner balance")
		fmt.Println("  5. help     - Show help")
		fmt.Println("  6. quit     - Exit")
		fmt.Print("\nEnter command: ")

		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch command {
		case "1", "mine":
			fmt.Print("Enter number of blocks to mine (default 1): ")
			if !scanner.Scan() {
				continue
			}
			numStr := strings.TrimSpace(scanner.Text())
			num := 1
			if numStr != "" {
				n, err := strconv.Atoi(numStr)
				if err == nil && n > 0 {
					num = n
				}
			}
			startMining(num)

		case "2", "chain":
			showBlockchain()

		case "3", "mempool":
			showMempoolContent()

		case "4", "balance":
			showMinerBalance()

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
	fmt.Println("Bitcoin Model Blockchain Miner")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  miner -createwallet              Create a new miner wallet")
	fmt.Println("  miner -init                      Initialize blockchain")
	fmt.Println("  miner -mine [-n=<num>]           Start mining (default 1 block)")
	fmt.Println("  miner -showchain                 Show blockchain")
	fmt.Println("  miner -showmempool               Show pending transactions")
	fmt.Println("  miner -balance                   Show miner balance")
	fmt.Println("  miner -interactive               Run in interactive mode")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  miner -init")
	fmt.Println("  miner -mine")
	fmt.Println("  miner -mine -n=5")
	fmt.Println("  miner -interactive")
}

// printInteractiveHelp 打印交互模式帮助
func printInteractiveHelp() {
	fmt.Println("\nHelp:")
	fmt.Println("  mine    - Mine new blocks")
	fmt.Println("            You will be prompted for the number of blocks")
	fmt.Println("  chain   - Display the full blockchain")
	fmt.Println("  mempool - Show pending transactions waiting to be mined")
	fmt.Println("  balance - Check your current balance")
	fmt.Println("  help    - Show this help message")
	fmt.Println("  quit    - Exit the program")
}
