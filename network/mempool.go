package network

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bitcoin-model-blockchain/core"
)

// Mempool 内存池（待处理交易池）
type Mempool struct {
	Transactions map[string]*core.Transaction
	mu           sync.RWMutex
	nodeID       string
}

// NewMempool 创建新内存池
func NewMempool(nodeID string) *Mempool {
	mp := &Mempool{
		Transactions: make(map[string]*core.Transaction),
		nodeID:       nodeID,
	}

	// 尝试加载已有的内存池
	mp.Load()

	return mp
}

// AddTransaction 添加交易到内存池
func (mp *Mempool) AddTransaction(tx *core.Transaction) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	txID := hex.EncodeToString(tx.ID)

	// 检查交易是否已存在
	if _, exists := mp.Transactions[txID]; exists {
		return fmt.Errorf("transaction already exists in mempool")
	}

	// 验证交易
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("transaction validation failed: %v", err)
	}

	mp.Transactions[txID] = tx

	// 保存到文件
	mp.save()

	fmt.Printf("Transaction %s added to mempool\n", txID)
	return nil
}

// GetTransaction 从内存池获取交易
func (mp *Mempool) GetTransaction(txID string) (*core.Transaction, bool) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	tx, exists := mp.Transactions[txID]
	return tx, exists
}

// GetTransactionsForBlock 获取用于打包区块的交易
func (mp *Mempool) GetTransactionsForBlock(maxCount int) []*core.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var transactions []*core.Transaction
	count := 0

	for _, tx := range mp.Transactions {
		if count >= maxCount {
			break
		}
		transactions = append(transactions, tx)
		count++
	}

	return transactions
}

// GetAllTransactions 获取所有待处理交易
func (mp *Mempool) GetAllTransactions() []*core.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var transactions []*core.Transaction
	for _, tx := range mp.Transactions {
		transactions = append(transactions, tx)
	}

	return transactions
}

// RemoveTransaction 从内存池移除交易
func (mp *Mempool) RemoveTransaction(txID []byte) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	txIDStr := hex.EncodeToString(txID)
	delete(mp.Transactions, txIDStr)

	// 保存到文件
	mp.save()
}

// RemoveTransactions 批量移除交易
func (mp *Mempool) RemoveTransactions(txIDs [][]byte) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, txID := range txIDs {
		txIDStr := hex.EncodeToString(txID)
		delete(mp.Transactions, txIDStr)
	}

	// 保存到文件
	mp.save()
}

// Contains 检查交易是否在内存池中
func (mp *Mempool) Contains(txID []byte) bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	txIDStr := hex.EncodeToString(txID)
	_, exists := mp.Transactions[txIDStr]
	return exists
}

// Count 获取内存池中交易数量
func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.Transactions)
}

// Clear 清空内存池
func (mp *Mempool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.Transactions = make(map[string]*core.Transaction)
	mp.save()
}

// CleanOldTransactions 清理旧交易（超过指定时间的交易）
func (mp *Mempool) CleanOldTransactions(maxAge time.Duration) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	now := time.Now().Unix()
	for txID, tx := range mp.Transactions {
		if now-tx.Timestamp > int64(maxAge.Seconds()) {
			delete(mp.Transactions, txID)
			fmt.Printf("Removed old transaction %s from mempool\n", txID)
		}
	}

	mp.save()
}

// save 保存内存池到文件
func (mp *Mempool) save() error {
	filename := mp.getMempoolFile()

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(mp.Transactions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// Load 从文件加载内存池
func (mp *Mempool) Load() error {
	filename := mp.getMempoolFile()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var transactions map[string]*core.Transaction
	if err := json.Unmarshal(data, &transactions); err != nil {
		return err
	}

	mp.Transactions = transactions
	return nil
}

// getMempoolFile 获取内存池文件路径
func (mp *Mempool) getMempoolFile() string {
	if mp.nodeID == "" {
		return filepath.Join("./data", "mempool.dat")
	}
	return filepath.Join("./data", fmt.Sprintf("mempool_%s.dat", mp.nodeID))
}

// String 返回内存池的字符串表示
func (mp *Mempool) String() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	result := fmt.Sprintf("Mempool (%d transactions):\n", len(mp.Transactions))
	for txID, tx := range mp.Transactions {
		result += fmt.Sprintf("  %s: %d inputs, %d outputs\n",
			txID, len(tx.Inputs), len(tx.Outputs))
	}
	return result
}

// BroadcastTransaction 广播交易到网络（简化版，写入共享文件）
func (mp *Mempool) BroadcastTransaction(tx *core.Transaction) error {
	// 在实际P2P网络中，这里会将交易广播给所有连接的节点
	// 简化版：我们将交易写入一个共享的广播文件
	filename := filepath.Join("./data", "broadcast_tx.dat")

	txData, err := tx.Serialize()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入交易长度和交易数据
	length := make([]byte, 4)
	length[0] = byte(len(txData) >> 24)
	length[1] = byte(len(txData) >> 16)
	length[2] = byte(len(txData) >> 8)
	length[3] = byte(len(txData))

	if _, err := file.Write(length); err != nil {
		return err
	}
	if _, err := file.Write(txData); err != nil {
		return err
	}

	return nil
}

// ReceiveBroadcastedTransactions 接收广播的交易
func (mp *Mempool) ReceiveBroadcastedTransactions() error {
	filename := filepath.Join("./data", "broadcast_tx.dat")

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	offset := 0
	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		length := int(data[offset])<<24 | int(data[offset+1])<<16 |
			int(data[offset+2])<<8 | int(data[offset+3])
		offset += 4

		if offset+length > len(data) {
			break
		}

		txData := data[offset : offset+length]
		offset += length

		tx, err := core.DeserializeTransaction(txData)
		if err != nil {
			continue
		}

		// 添加到内存池（如果不存在）
		txID := hex.EncodeToString(tx.ID)
		if _, exists := mp.Transactions[txID]; !exists {
			mp.Transactions[txID] = tx
		}
	}

	// 清空广播文件
	os.Remove(filename)

	return mp.save()
}

// bytesEqual 比较两个字节数组
func bytesEqual(a, b []byte) bool {
	return bytes.Equal(a, b)
}
