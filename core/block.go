package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"
)

// BlockHeader 区块头结构
type BlockHeader struct {
	PrevBlockHash []byte // 前一区块哈希
	MerkleRoot    []byte // Merkle树根哈希
	Timestamp     int64  // 时间戳
	Difficulty    uint32 // 难度目标
	Nonce         uint32 // 随机数
}

// Block 区块结构
type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
	Hash         []byte // 当前区块哈希（缓存）
	Height       uint32 // 区块高度
}

// NewBlock 创建新区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height uint32, difficulty uint32) *Block {
	block := &Block{
		Header: BlockHeader{
			PrevBlockHash: prevBlockHash,
			Timestamp:     time.Now().Unix(),
			Difficulty:    difficulty,
			Nonce:         0,
		},
		Transactions: transactions,
		Height:       height,
	}

	// 计算Merkle根
	block.Header.MerkleRoot = block.CalculateMerkleRoot()

	return block
}

// NewGenesisBlock 创建创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0, 16)
}

// CalculateMerkleRoot 计算Merkle树根哈希
func (b *Block) CalculateMerkleRoot() []byte {
	if len(b.Transactions) == 0 {
		return []byte{}
	}

	var hashes [][]byte
	for _, tx := range b.Transactions {
		hashes = append(hashes, tx.Hash)
	}

	// 如果交易数量为奇数，复制最后一个
	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var newLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			hash := sha256.Sum256(append(hashes[i], hashes[i+1]...))
			hash = sha256.Sum256(hash[:]) // 双重哈希
			newLevel = append(newLevel, hash[:])
		}
		hashes = newLevel
	}

	return hashes[0]
}

// SerializeHeader 序列化区块头（用于挖矿）
func (h *BlockHeader) SerializeHeader() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	encoder.Encode(h)
	return buf.Bytes()
}

// HashBlock 计算区块哈希（双重SHA256）
func (b *Block) HashBlock() []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	encoder.Encode(b.Header)
	hash := sha256.Sum256(buf.Bytes())
	hash = sha256.Sum256(hash[:])
	return hash[:]
}

// SetHash 设置区块哈希
func (b *Block) SetHash() {
	b.Hash = b.HashBlock()
}

// Serialize 序列化整个区块
func (b *Block) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeBlock 反序列化区块
func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// String 返回区块的字符串表示
func (b *Block) String() string {
	return fmt.Sprintf(
		"Block %d:\n  Hash: %s\n  PrevHash: %s\n  MerkleRoot: %s\n  Timestamp: %d\n  Difficulty: %d\n  Nonce: %d\n  Transactions: %d",
		b.Height,
		hex.EncodeToString(b.Hash),
		hex.EncodeToString(b.Header.PrevBlockHash),
		hex.EncodeToString(b.Header.MerkleRoot),
		b.Header.Timestamp,
		b.Header.Difficulty,
		b.Header.Nonce,
		len(b.Transactions),
	)
}

// Validate 验证区块有效性
func (b *Block) Validate() error {
	// 验证区块哈希
	if !bytes.Equal(b.Hash, b.HashBlock()) {
		return fmt.Errorf("invalid block hash")
	}

	// 验证Merkle根
	if !bytes.Equal(b.Header.MerkleRoot, b.CalculateMerkleRoot()) {
		return fmt.Errorf("invalid merkle root")
	}

	// 验证每个交易
	for _, tx := range b.Transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction: %v", err)
		}
	}

	return nil
}
