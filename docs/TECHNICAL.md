# 技术文档 - Bitcoin Model Blockchain

## 目录
1. [系统架构](#系统架构)
2. [核心组件](#核心组件)
3. [数据结构](#数据结构)
4. [密码学实现](#密码学实现)
5. [共识机制](#共识机制)
6. [交易流程](#交易流程)
7. [挖矿流程](#挖矿流程)
8. [存储机制](#存储机制)

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Client     │  │    Miner     │  │     CLI      │      │
│  │   (cmd/)     │  │    (cmd/)    │  │   Interface  │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼─────────────────┼─────────────────┼──────────────┘
          │                 │                 │
┌─────────┼─────────────────┼─────────────────┼──────────────┐
│         │    Core Layer   │                 │              │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐      │
│  │   Block      │  │ Transaction  │  │    Wallet    │      │
│  │   (block)    │  │    (tx)      │  │  (crypto)    │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │              │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐      │
│  │ Blockchain   │  │    UTXO      │  │    Miner     │      │
│  │   (chain)    │  │    (set)     │  │   (engine)   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
          │                 │                 │
┌─────────┼─────────────────┼─────────────────┼──────────────┐
│         │   Network Layer │                 │              │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐      │
│  │   Mempool    │  │   Broadcast  │  │   Storage    │      │
│  │  (pending)   │  │   (simplified)│  │  (persist)   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

---

## 核心组件

### 1. Block (区块)

**文件**: `core/block.go`

区块是区块链的基本组成单元，包含区块头和交易列表。

**区块头结构**:
```go
type BlockHeader struct {
    PrevBlockHash []byte // 前一区块哈希 (32 bytes)
    MerkleRoot    []byte // Merkle树根哈希 (32 bytes)
    Timestamp     int64  // 时间戳 (8 bytes)
    Difficulty    uint32 // 难度目标 (4 bytes)
    Nonce         uint32 // 随机数 (4 bytes)
}
```

**区块结构**:
```go
type Block struct {
    Header       BlockHeader
    Transactions []*Transaction
    Hash         []byte
    Height       uint32
}
```

**关键方法**:
- `NewBlock()`: 创建新区块
- `CalculateMerkleRoot()`: 计算Merkle根
- `HashBlock()`: 计算区块哈希（双重SHA256）
- `Validate()`: 验证区块有效性

### 2. Transaction (交易)

**文件**: `core/transaction.go`

交易是价值转移的载体，采用UTXO模型。

**交易输入**:
```go
type TXInput struct {
    TxID      []byte // 引用的交易ID
    Vout      int    // 引用的输出索引
    Signature []byte // 签名 (64 bytes)
    PubKey    []byte // 公钥
}
```

**交易输出**:
```go
type TXOutput struct {
    Value      int    // 金额
    PubKeyHash []byte // 接收方公钥哈希 (20 bytes)
}
```

**交易类型**:
- **Coinbase交易**: 区块奖励，无输入，Vout = -1
- **普通交易**: 转账交易，引用UTXO作为输入

### 3. Wallet (钱包)

**文件**: `core/wallet.go`

钱包管理密钥对，使用ECDSA (secp256k1) 椭圆曲线加密。

**钱包结构**:
```go
type Wallet struct {
    PrivateKey ecdsa.PrivateKey
    PublicKey  []byte
    Address    string
}
```

**地址生成流程**:
1. 生成ECDSA密钥对
2. 提取公钥 (X, Y坐标拼接)
3. SHA256哈希
4. 取前20字节作为PubKeyHash
5. 添加版本字节和校验和
6. Base58编码 (简化版使用Hex)

### 4. Blockchain (区块链)

**文件**: `core/blockchain.go`

管理区块的链式结构，提供验证和查询功能。

**核心功能**:
- 区块验证（哈希、高度、前一区块）
- 交易验证（签名、UTXO、双花检查）
- 持久化存储（JSON格式）

### 5. UTXOSet (UTXO集合)

**文件**: `core/utxo.go`

管理未花费交易输出，是余额计算的基础。

```go
type UTXOSet struct {
    UTXO map[string][]TXOutput  // TxID -> Outputs
}
```

**关键方法**:
- `FindUTXO()`: 查找地址的所有UTXO
- `GetBalance()`: 计算地址余额
- `Update()`: 用新区块更新UTXO集合
- `Reindex()`: 从区块链重建UTXO集合

### 6. ProofOfWork (工作量证明)

**文件**: `core/miner.go`

实现PoW共识机制。

**挖矿算法**:
```
目标值 = 2^(256 - difficulty)
循环:
    数据 = PrevHash + MerkleRoot + Timestamp + Difficulty + Nonce
    哈希 = SHA256(SHA256(数据))
    如果 哈希 < 目标值:
        找到有效Nonce
    Nonce++
```

**难度调整**:
- 初始难度: 16位 (4个前导0)
- 简化版固定难度，实际比特币每2016区块调整

---

## 数据结构

### Merkle树

Merkle树用于高效验证交易完整性。

**构建过程**:
```
Level 0: [Tx1] [Tx2] [Tx3] [Tx4]
           ↓     ↓     ↓     ↓
Level 1: [Hash1-2]   [Hash3-4]
              ↓          ↓
Level 2:   [MerkleRoot]
```

**特性**:
- 二叉树结构
- 叶子节点是交易哈希
- 父节点是子节点拼接后的双重SHA256
- 奇数节点时复制最后一个

### 区块链存储格式

**文件**: `data/blockchain_<nodeID>.dat`

```json
[
  {
    "Header": {
      "PrevBlockHash": "base64",
      "MerkleRoot": "base64",
      "Timestamp": 1234567890,
      "Difficulty": 16,
      "Nonce": 12345
    },
    "Transactions": [...],
    "Hash": "base64",
    "Height": 0
  }
]
```

---

## 密码学实现

### 1. 哈希算法

**双重SHA256**:
```go
func sha256DoubleHash(data []byte) [32]byte {
    firstHash := sha256.Sum256(data)
    return sha256.Sum256(firstHash[:])
}
```

### 2. 数字签名

**ECDSA (secp256k1)**:

签名过程:
```go
func SignData(privKey []byte, data []byte) ([]byte, error) {
    privateKey, _ := x509.ParseECPrivateKey(privKey)
    r, s, _ := ecdsa.Sign(rand.Reader, privateKey, data)
    signature := append(r.Bytes(), s.Bytes()...)
    return signature, nil
}
```

验证过程:
```go
func VerifySignature(pubKey []byte, data []byte, signature []byte) bool {
    // 解析公钥和签名
    // 调用 ecdsa.Verify()
}
```

### 3. 交易签名流程

1. 创建交易副本（清除输入的Signature和PubKey）
2. 对每个输入，设置PubKey为引用的输出的PubKeyHash
3. 计算交易哈希
4. 使用私钥签名哈希
5. 将签名保存到原始交易的输入中

---

## 共识机制

### PoW (Proof of Work)

**目标**: 找到Nonce使得 `SHA256(SHA256(区块头)) < 目标值`

**难度计算**:
```
目标值 = 2^(256 - difficulty)
```

**示例** (difficulty=16):
- 目标值 = 2^240
- 有效哈希必须以4个十六进制0开头 (如: 0000abcd...)

**安全性**:
- 计算困难，验证容易
- 51%攻击需要控制大部分算力
- 最长链原则

---

## 交易流程

### 创建交易

```
1. 查询发送方UTXO
   ↓
2. 选择足够金额的UTXO
   ↓
3. 构建输入（引用UTXO）
   ↓
4. 构建输出（收款方+找零）
   ↓
5. 签名交易
   ↓
6. 广播到网络
```

### 验证交易

```
1. 检查交易格式
   ↓
2. 验证输入引用的UTXO存在
   ↓
3. 验证签名
   ↓
4. 检查输入金额 >= 输出金额
   ↓
5. 检查无双重花费
```

---

## 挖矿流程

```
1. 从Mempool获取待处理交易
   ↓
2. 创建Coinbase交易（奖励）
   ↓
3. 构建候选区块
   ↓
4. 计算Merkle根
   ↓
5. 执行PoW挖矿
   ↓
6. 找到有效Nonce
   ↓
7. 验证区块
   ↓
8. 添加到区块链
   ↓
9. 更新UTXO集合
   ↓
10. 从Mempool移除已打包交易
```

---

## 存储机制

### 数据文件

| 文件 | 内容 | 格式 |
|------|------|------|
| `blockchain_*.dat` | 区块数据 | JSON |
| `mempool_*.dat` | 待处理交易 | JSON |
| `utxo_*.dat` | UTXO集合 | JSON |
| `wallet_*.dat` | 钱包密钥 | JSON |
| `broadcast_tx.dat` | 广播交易 | Binary |

### 序列化

使用Go的`encoding/gob`进行二进制序列化，JSON用于持久化存储便于调试。

---

## 安全考虑

1. **私钥存储**: 明文JSON存储，生产环境应加密
2. **交易验证**: 完整的签名和UTXO验证
3. **双花防护**: UTXO模型天然防止双花
4. **难度调整**: 固定难度简化，实际应动态调整

---

## 性能优化点

1. **Merkle树**: O(log n) 交易验证
2. **UTXO缓存**: 内存中维护UTXO集合
3. **批量验证**: 区块级别批量验证交易
4. **难度控制**: 可调节挖矿难度适应硬件

---

## 扩展性

### 可添加功能

1. **P2P网络**: 实现真正的节点通信
2. **难度调整**: 动态难度算法
3. **脚本系统**: 更复杂的锁定/解锁脚本
4. **轻节点**: SPV验证支持
5. **分片**: 提高吞吐量
