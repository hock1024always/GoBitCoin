package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// CoinbaseReward 挖矿奖励
	CoinbaseReward = 50
	// CoinbaseData 创世区块标识
	CoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

// TXInput 交易输入
type TXInput struct {
	TxID      []byte // 引用的交易ID
	Vout      int    // 引用的输出索引
	Signature []byte // 签名
	PubKey    []byte // 公钥（解锁脚本）
}

// TXOutput 交易输出
type TXOutput struct {
	Value      int    // 金额
	PubKeyHash []byte // 接收方公钥哈希（锁定脚本）
}

// Transaction 交易结构
type Transaction struct {
	ID        []byte     // 交易哈希
	Hash      []byte     // 交易哈希（别名，用于兼容）
	Inputs    []TXInput  // 输入列表
	Outputs   []TXOutput // 输出列表
	Timestamp int64      // 时间戳
}

// NewCoinbaseTX 创建创世区块奖励交易
func NewCoinbaseTX(to string, data string) *Transaction {
	if data == "" {
		data = CoinbaseData
	}

	// Coinbase交易没有输入，TxID为空，Vout为-1
	txIn := TXInput{
		TxID:      []byte{},
		Vout:      -1,
		Signature: nil,
		PubKey:    []byte(data),
	}

	txOut := TXOutput{
		Value:      CoinbaseReward,
		PubKeyHash: GetPubKeyHashFromAddress(to),
	}

	tx := &Transaction{
		Inputs:    []TXInput{txIn},
		Outputs:   []TXOutput{txOut},
		Timestamp: time.Now().Unix(),
	}
	tx.SetID()

	return tx
}

// NewUTXOTransaction 创建普通转账交易
func NewUTXOTransaction(from string, to string, amount int, utxoSet map[string][]TXOutput, privKey []byte, pubKey []byte) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput

	// 查找可用的UTXO
	acc, validOutputs := FindSpendableOutputs(utxoSet, from, amount)

	if acc < amount {
		return nil, fmt.Errorf("insufficient funds: have %d, need %d", acc, amount)
	}

	// 构建输入
	for txID, outs := range validOutputs {
		txIDBytes, _ := hex.DecodeString(txID)

		for _, out := range outs {
			input := TXInput{
				TxID:   txIDBytes,
				Vout:   out,
				PubKey: pubKey,
			}
			inputs = append(inputs, input)
		}
	}

	// 构建输出
	outputs = append(outputs, TXOutput{
		Value:      amount,
		PubKeyHash: HashPubKey([]byte(to)),
	})

	// 找零
	if acc > amount {
		outputs = append(outputs, TXOutput{
			Value:      acc - amount,
			PubKeyHash: HashPubKey([]byte(from)),
		})
	}

	tx := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now().Unix(),
	}
	tx.SetID()

	// 签名交易
	err := tx.Sign(privKey, utxoSet)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// SetID 计算并设置交易ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	encoder.Encode(tx.Inputs)
	encoder.Encode(tx.Outputs)
	encoder.Encode(tx.Timestamp)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
	tx.Hash = hash[:] // 同时设置Hash字段
}

// IsCoinbase 判断是否为Coinbase交易
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TxID) == 0 && tx.Inputs[0].Vout == -1
}

// Serialize 序列化交易
func (tx *Transaction) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(tx)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeTransaction 反序列化交易
func DeserializeTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// Sign 签名交易
func (tx *Transaction) Sign(privKey []byte, utxoSet map[string][]TXOutput) error {
	if tx.IsCoinbase() {
		return nil
	}

	// 创建交易的副本用于签名（清除输入的签名和公钥）
	txCopy := tx.TrimmedCopy()

	for inID, input := range txCopy.Inputs {
		// 获取引用的输出
		prevTXID := hex.EncodeToString(input.TxID)
		prevOutputs, exists := utxoSet[prevTXID]
		if !exists || input.Vout >= len(prevOutputs) {
			return fmt.Errorf("referenced output not found")
		}

		prevOutput := prevOutputs[input.Vout]
		txCopy.Inputs[inID].PubKey = prevOutput.PubKeyHash
		txCopy.SetID()
		txCopy.Inputs[inID].PubKey = nil

		// 使用私钥签名
		signature, err := SignData(privKey, txCopy.ID)
		if err != nil {
			return err
		}

		tx.Inputs[inID].Signature = signature
	}

	return nil
}

// Verify 验证交易签名
func (tx *Transaction) Verify(utxoSet map[string][]TXOutput) error {
	if tx.IsCoinbase() {
		return nil
	}

	txCopy := tx.TrimmedCopy()

	for inID, input := range tx.Inputs {
		prevTXID := hex.EncodeToString(input.TxID)
		prevOutputs, exists := utxoSet[prevTXID]
		if !exists || input.Vout >= len(prevOutputs) {
			return fmt.Errorf("referenced output not found")
		}

		prevOutput := prevOutputs[input.Vout]
		txCopy.Inputs[inID].PubKey = prevOutput.PubKeyHash
		txCopy.SetID()
		txCopy.Inputs[inID].PubKey = nil

		// 验证签名
		if !VerifySignature(input.PubKey, txCopy.ID, input.Signature) {
			return fmt.Errorf("invalid signature for input %d", inID)
		}
	}

	return nil
}

// TrimmedCopy 创建用于签名的交易副本
func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TXInput{
			TxID:      in.TxID,
			Vout:      in.Vout,
			Signature: nil,
			PubKey:    nil,
		})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TXOutput{
			Value:      out.Value,
			PubKeyHash: out.PubKeyHash,
		})
	}

	return &Transaction{
		ID:        tx.ID,
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: tx.Timestamp,
	}
}

// Validate 验证交易有效性
func (tx *Transaction) Validate() error {
	if len(tx.ID) == 0 {
		return fmt.Errorf("transaction ID is empty")
	}

	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	for _, out := range tx.Outputs {
		if out.Value <= 0 {
			return fmt.Errorf("invalid output value: %d", out.Value)
		}
	}

	return nil
}

// String 返回交易的字符串表示
func (tx *Transaction) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Transaction %s:\n", hex.EncodeToString(tx.ID)))
	buf.WriteString(fmt.Sprintf("  Timestamp: %d\n", tx.Timestamp))
	buf.WriteString(fmt.Sprintf("  Inputs (%d):\n", len(tx.Inputs)))
	for i, in := range tx.Inputs {
		buf.WriteString(fmt.Sprintf("    [%d] TxID: %s, Vout: %d\n",
			i, hex.EncodeToString(in.TxID), in.Vout))
	}
	buf.WriteString(fmt.Sprintf("  Outputs (%d):\n", len(tx.Outputs)))
	for i, out := range tx.Outputs {
		buf.WriteString(fmt.Sprintf("    [%d] Value: %d, PubKeyHash: %s\n",
			i, out.Value, hex.EncodeToString(out.PubKeyHash)))
	}

	return buf.String()
}
