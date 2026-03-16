package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const utxoFile = "utxo.dat"

// UTXOSet UTXO集合管理
type UTXOSet struct {
	// map[交易ID] -> []TXOutput
	UTXO map[string][]TXOutput
}

// NewUTXOSet 创建新的UTXO集合
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		UTXO: make(map[string][]TXOutput),
	}
}

// FindSpendableOutputs 查找可花费的UTXO
func FindSpendableOutputs(utxoSet map[string][]TXOutput, address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	targetPubKeyHash := GetPubKeyHashFromAddress(address)

Work:
	for txID, outputs := range utxoSet {
		for outIdx, out := range outputs {
			if accumulated >= amount {
				break Work
			}

			if out.IsLockedWithKey(targetPubKeyHash) {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}
		}
	}

	return accumulated, unspentOutputs
}

// FindUTXO 查找地址的所有UTXO
func (u *UTXOSet) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	// 从地址提取PubKeyHash（地址格式: version(1) + PubKeyHash(20) + checksum(4)）
	// 简化版：地址就是PubKeyHash的hex编码
	targetPubKeyHash := GetPubKeyHashFromAddress(address)

	for _, outputs := range u.UTXO {
		for _, out := range outputs {
			if out.IsLockedWithKey(targetPubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// GetBalance 获取地址余额
func (u *UTXOSet) GetBalance(address string) int {
	utxos := u.FindUTXO(address)
	balance := 0

	for _, out := range utxos {
		balance += out.Value
	}

	return balance
}

// Update 用新区块更新UTXO集合
func (u *UTXOSet) Update(block *Block) {
	for _, tx := range block.Transactions {
		// 如果不是Coinbase交易，删除已花费的输入
		if !tx.IsCoinbase() {
			for _, in := range tx.Inputs {
				txID := hex.EncodeToString(in.TxID)
				if outputs, exists := u.UTXO[txID]; exists {
					// 删除已花费的输出
					if in.Vout < len(outputs) {
						u.UTXO[txID] = append(outputs[:in.Vout], outputs[in.Vout+1:]...)
						// 如果该交易的所有输出都被花费，删除该交易
						if len(u.UTXO[txID]) == 0 {
							delete(u.UTXO, txID)
						}
					}
				}
			}
		}

		// 添加新的输出
		txID := hex.EncodeToString(tx.ID)
		u.UTXO[txID] = tx.Outputs
	}
}

// Reindex 从区块链重新构建UTXO集合
func (u *UTXOSet) Reindex(blocks []*Block) {
	u.UTXO = make(map[string][]TXOutput)

	for _, block := range blocks {
		u.Update(block)
	}
}

// SaveToFile 保存UTXO集合到文件
func (u *UTXOSet) SaveToFile(nodeID string) error {
	filename := getUTXOFile(nodeID)

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(u.UTXO, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadUTXOSet 从文件加载UTXO集合
func LoadUTXOSet(nodeID string) (*UTXOSet, error) {
	filename := getUTXOFile(nodeID)

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return NewUTXOSet(), nil
		}
		return nil, err
	}

	var utxoMap map[string][]TXOutput
	if err := json.Unmarshal(data, &utxoMap); err != nil {
		return nil, err
	}

	return &UTXOSet{UTXO: utxoMap}, nil
}

// getUTXOFile 获取UTXO文件路径
func getUTXOFile(nodeID string) string {
	if nodeID == "" {
		return filepath.Join("./data", utxoFile)
	}
	return filepath.Join("./data", fmt.Sprintf("utxo_%s.dat", nodeID))
}

// IsLockedWithKey 检查输出是否被指定公钥哈希锁定
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return string(out.PubKeyHash) == string(pubKeyHash)
}

// String 返回UTXO的字符串表示
func (u *UTXOSet) String() string {
	var result string
	result = fmt.Sprintf("UTXO Set (%d transactions):\n", len(u.UTXO))

	for txID, outputs := range u.UTXO {
		result += fmt.Sprintf("  TxID: %s\n", txID)
		for i, out := range outputs {
			result += fmt.Sprintf("    [%d] Value: %d, PubKeyHash: %s\n",
				i, out.Value, hex.EncodeToString(out.PubKeyHash))
		}
	}

	return result
}
