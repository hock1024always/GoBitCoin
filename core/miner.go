package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"time"
)

const (
	// MaxNonce 最大Nonce值
	MaxNonce = math.MaxInt32
	// TargetBits 初始难度（前导0的位数）
	TargetBits = 16
)

// ProofOfWork 工作量证明结构
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork 创建新的工作量证明
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-b.Header.Difficulty))

	return &ProofOfWork{
		block:  b,
		target: target,
	}
}

// prepareData 准备挖矿数据
func (pow *ProofOfWork) prepareData(nonce uint32) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.Header.PrevBlockHash,
			pow.block.Header.MerkleRoot,
			IntToHex(pow.block.Header.Timestamp),
			IntToHex(int64(pow.block.Header.Difficulty)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

// Run 执行挖矿
func (pow *ProofOfWork) Run() (uint32, []byte) {
	var hashInt big.Int
	var hash [32]byte
	var nonce uint32

	fmt.Printf("Mining block with target difficulty: %d\n", pow.block.Header.Difficulty)
	startTime := time.Now()

	for nonce < MaxNonce {
		data := pow.prepareData(nonce)
		hash = sha256DoubleHash(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			duration := time.Since(startTime)
			fmt.Printf("Block mined! Nonce: %d, Hash: %x\n", nonce, hash)
			fmt.Printf("Time elapsed: %v, Hash rate: %.2f H/s\n",
				duration, float64(nonce)/duration.Seconds())
			return nonce, hash[:]
		}

		nonce++

		// 每100万次哈希打印一次进度
		if nonce%1000000 == 0 {
			fmt.Printf("Mining... nonce: %d\r", nonce)
		}
	}

	fmt.Println("\nMining failed: reached max nonce")
	return 0, nil
}

// Validate 验证区块的工作量证明
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Header.Nonce)
	hash := sha256DoubleHash(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}

// sha256DoubleHash 双重SHA256哈希
func sha256DoubleHash(data []byte) [32]byte {
	firstHash := sha256.Sum256(data)
	return sha256.Sum256(firstHash[:])
}



// IntToHex 将int64转换为字节数组
func IntToHex(num int64) []byte {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, uint64(num))
	return buff
}

// Miner 矿工结构
type Miner struct {
	Address      string
	Blockchain   *Blockchain
	Mempool      MempoolInterface
	UTXOSet      *UTXOSet
	MiningReward int
}

// MempoolInterface 内存池接口
type MempoolInterface interface {
	GetTransactionsForBlock(maxCount int) []*Transaction
	RemoveTransaction(txID []byte)
}

// NewMiner 创建新矿工
func NewMiner(address string, bc *Blockchain, mempool MempoolInterface, utxoSet *UTXOSet) *Miner {
	return &Miner{
		Address:      address,
		Blockchain:   bc,
		Mempool:      mempool,
		UTXOSet:      utxoSet,
		MiningReward: CoinbaseReward,
	}
}

// MineBlock 挖矿生成新区块
func (m *Miner) MineBlock() (*Block, error) {
	// 获取待处理交易
	transactions := m.Mempool.GetTransactionsForBlock(10) // 最多10笔交易

	// 如果没有交易，返回错误
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions to mine")
	}

	// 添加Coinbase交易
	coinbaseTX := NewCoinbaseTX(m.Address, "")
	transactions = append([]*Transaction{coinbaseTX}, transactions...)

	// 验证所有交易
	for _, tx := range transactions {
		if !tx.IsCoinbase() {
			if err := m.Blockchain.VerifyTransaction(tx); err != nil {
				fmt.Printf("Invalid transaction: %v, removing from mempool\n", err)
				m.Mempool.RemoveTransaction(tx.ID)
				continue
			}
		}
	}

	// 获取最新区块
	prevBlock := m.Blockchain.GetLatestBlock()

	// 计算难度（简化版，固定难度）
	difficulty := uint32(TargetBits)

	// 创建新区块
	newBlock := NewBlock(
		transactions,
		prevBlock.Hash,
		prevBlock.Height+1,
		difficulty,
	)

	// 执行工作量证明
	pow := NewProofOfWork(newBlock)
	nonce, hash := pow.Run()

	if hash == nil {
		return nil, fmt.Errorf("mining failed")
	}

	newBlock.Header.Nonce = nonce
	newBlock.SetHash()

	// 验证区块
	if err := newBlock.Validate(); err != nil {
		return nil, fmt.Errorf("block validation failed: %v", err)
	}

	// 添加区块到区块链
	if err := m.Blockchain.AddBlock(newBlock); err != nil {
		return nil, err
	}

	// 更新UTXO集合
	m.UTXOSet.Update(newBlock)

	// 从内存池移除已打包的交易
	for _, tx := range transactions {
		if !tx.IsCoinbase() {
			m.Mempool.RemoveTransaction(tx.ID)
		}
	}

	fmt.Printf("Successfully mined block #%d with %d transactions\n",
		newBlock.Height, len(transactions))

	return newBlock, nil
}

// CalculateDifficulty 计算挖矿难度（简化版）
func CalculateDifficulty(blockchain *Blockchain) uint32 {
	// 实际比特币每2016个区块调整一次难度
	// 这里简化，返回固定难度
	return TargetBits
}

// HashRate 计算哈希率
func HashRate(startTime time.Time, nonce uint32) float64 {
	duration := time.Since(startTime).Seconds()
	if duration == 0 {
		return 0
	}
	return float64(nonce) / duration
}
