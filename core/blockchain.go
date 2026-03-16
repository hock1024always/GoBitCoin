package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Blockchain 区块链结构
type Blockchain struct {
	Blocks   []*Block
	Tip      []byte // 最新区块哈希
	UTXOSet  *UTXOSet
	nodeID   string
}

// NewBlockchain 创建新区块链
func NewBlockchain(address string, nodeID string) (*Blockchain, error) {
	bc := &Blockchain{
		Blocks: make([]*Block, 0),
		Tip:    []byte{},
		UTXOSet: NewUTXOSet(),
		nodeID: nodeID,
	}

	// 检查是否存在已有区块链
	if bc.Exists() {
		return bc, bc.Load()
	}

	// 创建创世区块
	coinbaseTX := NewCoinbaseTX(address, "")
	genesis := NewGenesisBlock(coinbaseTX)
	genesis.SetHash()

	// 添加创世区块
	bc.Blocks = append(bc.Blocks, genesis)
	bc.Tip = genesis.Hash

	// 更新UTXO集合
	bc.UTXOSet.Update(genesis)

	// 保存到文件
	if err := bc.Save(); err != nil {
		return nil, err
	}

	return bc, nil
}

// LoadBlockchain 加载已有区块链
func LoadBlockchain(nodeID string) (*Blockchain, error) {
	bc := &Blockchain{
		Blocks: make([]*Block, 0),
		Tip:    []byte{},
		UTXOSet: NewUTXOSet(),
		nodeID: nodeID,
	}

	if err := bc.Load(); err != nil {
		return nil, err
	}

	return bc, nil
}

// AddBlock 添加新区块到区块链
func (bc *Blockchain) AddBlock(block *Block) error {
	// 验证区块
	if err := bc.ValidateBlock(block); err != nil {
		return err
	}

	// 添加区块
	bc.Blocks = append(bc.Blocks, block)
	bc.Tip = block.Hash

	// 保存
	return bc.Save()
}

// ValidateBlock 验证区块
func (bc *Blockchain) ValidateBlock(block *Block) error {
	// 验证区块哈希
	if !bytesEqual(block.Hash, block.HashBlock()) {
		return fmt.Errorf("invalid block hash")
	}

	// 验证前一区块哈希
	if len(bc.Blocks) > 0 {
		prevBlock := bc.Blocks[len(bc.Blocks)-1]
		if !bytesEqual(block.Header.PrevBlockHash, prevBlock.Hash) {
			return fmt.Errorf("invalid previous block hash")
		}

		// 验证区块高度
		if block.Height != prevBlock.Height+1 {
			return fmt.Errorf("invalid block height")
		}
	}

	// 验证工作量证明
	pow := NewProofOfWork(block)
	if !pow.Validate() {
		return fmt.Errorf("proof of work validation failed")
	}

	// 验证所有交易
	for _, tx := range block.Transactions {
		if err := bc.VerifyTransaction(tx); err != nil {
			return fmt.Errorf("transaction verification failed: %v", err)
		}
	}

	return nil
}

// VerifyTransaction 验证交易
func (bc *Blockchain) VerifyTransaction(tx *Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	// 验证交易签名
	if err := tx.Verify(bc.UTXOSet.UTXO); err != nil {
		return err
	}

	// 验证输入是否可用
	for _, in := range tx.Inputs {
		txID := hex.EncodeToString(in.TxID)
		outputs, exists := bc.UTXOSet.UTXO[txID]
		if !exists {
			return fmt.Errorf("referenced transaction not found: %s", txID)
		}
		if in.Vout >= len(outputs) {
			return fmt.Errorf("invalid output index")
		}
	}

	// 验证输入金额 >= 输出金额
	inputSum := 0
	for _, in := range tx.Inputs {
		txID := hex.EncodeToString(in.TxID)
		outputs := bc.UTXOSet.UTXO[txID]
		inputSum += outputs[in.Vout].Value
	}

	outputSum := 0
	for _, out := range tx.Outputs {
		outputSum += out.Value
	}

	if inputSum < outputSum {
		return fmt.Errorf("insufficient input value")
	}

	return nil
}

// GetLatestBlock 获取最新区块
func (bc *Blockchain) GetLatestBlock() *Block {
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

// GetBlockByHeight 根据高度获取区块
func (bc *Blockchain) GetBlockByHeight(height uint32) *Block {
	for _, block := range bc.Blocks {
		if block.Height == height {
			return block
		}
	}
	return nil
}

// GetBlockByHash 根据哈希获取区块
func (bc *Blockchain) GetBlockByHash(hash []byte) *Block {
	for _, block := range bc.Blocks {
		if bytesEqual(block.Hash, hash) {
			return block
		}
	}
	return nil
}

// GetHeight 获取区块链高度
func (bc *Blockchain) GetHeight() uint32 {
	if len(bc.Blocks) == 0 {
		return 0
	}
	return bc.Blocks[len(bc.Blocks)-1].Height
}

// Exists 检查区块链是否存在
func (bc *Blockchain) Exists() bool {
	filename := bc.getBlockchainFile()
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// Save 保存区块链到文件
func (bc *Blockchain) Save() error {
	filename := bc.getBlockchainFile()

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(bc.Blocks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// Load 从文件加载区块链
func (bc *Blockchain) Load() error {
	filename := bc.getBlockchainFile()

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &bc.Blocks); err != nil {
		return err
	}

	if len(bc.Blocks) > 0 {
		bc.Tip = bc.Blocks[len(bc.Blocks)-1].Hash
		// 重建UTXO集合
		bc.UTXOSet.Reindex(bc.Blocks)
	}

	return nil
}

// getBlockchainFile 获取区块链文件路径
func (bc *Blockchain) getBlockchainFile() string {
	if bc.nodeID == "" {
		return filepath.Join("./data", "blockchain.dat")
	}
	return filepath.Join("./data", fmt.Sprintf("blockchain_%s.dat", bc.nodeID))
}

// String 返回区块链的字符串表示
func (bc *Blockchain) String() string {
	result := fmt.Sprintf("Blockchain (Height: %d, Blocks: %d):\n", bc.GetHeight(), len(bc.Blocks))
	for _, block := range bc.Blocks {
		result += block.String() + "\n"
	}
	return result
}

// bytesEqual 比较两个字节数组
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
