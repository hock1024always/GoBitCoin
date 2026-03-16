# Bitcoin Model Blockchain

一个用Go语言实现的简化版比特币模型区块链，包含完整的区块结构、挖矿、交易和钱包功能。

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 特性

- **完整区块结构**: 区块头、Merkle树、双重SHA256哈希
- **PoW共识机制**: 工作量证明挖矿，可调节难度
- **UTXO模型**: 未花费交易输出，支持余额查询
- **数字签名**: ECDSA椭圆曲线加密，secp256k1曲线
- **交易系统**: 支持转账、找零、Coinbase奖励
- **钱包管理**: 密钥对生成、地址派生
- **持久化存储**: JSON格式存储区块链数据
- **双端CLI**: 独立的客户端和矿工端

## 项目结构

```
bitcoin-model-blockchain/
├── cmd/
│   ├── client/          # 客户端 - 发起交易、查询余额
│   │   └── main.go
│   └── miner/           # 矿工端 - 挖矿、管理区块链
│       └── main.go
├── core/                # 核心区块链逻辑
│   ├── block.go         # 区块结构
│   ├── transaction.go   # 交易结构
│   ├── blockchain.go    # 区块链管理
│   ├── wallet.go        # 钱包和加密
│   ├── utxo.go          # UTXO集合
│   └── miner.go         # 挖矿引擎
├── network/             # 网络层（简化版）
│   └── mempool.go       # 交易内存池
├── docs/                # 文档
│   └── TECHNICAL.md     # 技术文档
├── data/                # 数据存储目录（运行时创建）
├── wallets/             # 钱包存储目录（运行时创建）
├── go.mod
└── README.md
```

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- Linux/macOS/Windows

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/bitcoin-model-blockchain.git
cd bitcoin-model-blockchain

# 下载依赖
go mod tidy

# 编译
mkdir -p bin
go build -o bin/client ./cmd/client
go build -o bin/miner ./cmd/miner
```

### 运行演示

#### 1. 初始化矿工和区块链

```bash
# 创建矿工钱包
./bin/miner -createwallet

# 初始化区块链（创建创世区块）
./bin/miner -init
```

#### 2. 创建客户端钱包

```bash
# 创建客户端钱包
./bin/client -createwallet
```

#### 3. 矿工挖矿（获得奖励）

```bash
# 挖一个区块（包含Coinbase奖励交易）
./bin/miner -mine

# 查看矿工余额
./bin/miner -balance
```

#### 4. 客户端发起交易

```bash
# 查看客户端地址
./bin/client -listwallets

# 发送交易（假设矿工地址是 abc123...，客户端地址是 def456...）
./bin/client -address=def456... -to=abc123... -amount=10

# 查看客户端余额
./bin/client -balance -address=def456...
```

#### 5. 矿工打包交易

```bash
# 查看待处理交易
./bin/miner -showmempool

# 挖矿打包交易
./bin/miner -mine

# 查看更新后的余额
./bin/miner -balance
```

### 交互模式

```bash
# 启动客户端交互模式
./bin/client -interactive

# 启动矿工交互模式
./bin/miner -interactive
```

## 命令参考

### 客户端命令

| 命令 | 说明 |
|------|------|
| `-createwallet` | 创建新钱包 |
| `-listwallets` | 列出所有钱包 |
| `-balance -address=<addr>` | 查询余额 |
| `-address=<addr> -to=<addr> -amount=<n>` | 发送交易 |
| `-interactive` | 交互模式 |

### 矿工命令

| 命令 | 说明 |
|------|------|
| `-createwallet` | 创建矿工钱包 |
| `-init` | 初始化区块链 |
| `-mine [-n=<num>]` | 挖矿（默认1个区块） |
| `-showchain` | 显示区块链 |
| `-showmempool` | 显示待处理交易 |
| `-balance` | 显示矿工余额 |
| `-interactive` | 交互模式 |

## 技术亮点

### 1. 区块结构

```go
type Block struct {
    Header       BlockHeader      // 区块头
    Transactions []*Transaction   // 交易列表
    Hash         []byte          // 区块哈希
    Height       uint32          // 区块高度
}

type BlockHeader struct {
    PrevBlockHash []byte  // 前一区块哈希
    MerkleRoot    []byte  // Merkle树根
    Timestamp     int64   // 时间戳
    Difficulty    uint32  // 难度目标
    Nonce         uint32  // 随机数
}
```

### 2. 挖矿算法

使用双重SHA256进行工作量证明：

```
目标值 = 2^(256 - difficulty)

循环:
    数据 = PrevHash + MerkleRoot + Timestamp + Difficulty + Nonce
    哈希 = SHA256(SHA256(数据))
    如果 哈希 < 目标值:
        找到有效Nonce
    Nonce++
```

### 3. UTXO模型

```go
type TXInput struct {
    TxID      []byte  // 引用的交易ID
    Vout      int     // 引用的输出索引
    Signature []byte  // 签名
    PubKey    []byte  // 公钥
}

type TXOutput struct {
    Value      int    // 金额
    PubKeyHash []byte // 接收方公钥哈希
}
```

### 4. 数字签名

使用ECDSA (secp256k1) 进行签名和验证：

```go
// 签名
signature, _ := SignData(privateKey, transactionHash)

// 验证
valid := VerifySignature(publicKey, transactionHash, signature)
```

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Client     │  │    Miner     │  │     CLI      │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼─────────────────┼─────────────────┼──────────────┘
          │                 │                 │
┌─────────┼─────────────────┼─────────────────┼──────────────┐
│         │    Core Layer   │                 │              │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐      │
│  │   Block      │  │ Transaction  │  │    Wallet    │      │
│  │  (PoW/Merkle)│  │   (UTXO)     │  │  (ECDSA)     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## 详细文档

- [技术文档](docs/TECHNICAL.md) - 详细的技术实现说明

## 测试

```bash
# 运行单元测试
go test ./...

# 运行特定包测试
go test ./core -v
```

## 开发计划

- [x] 基础区块结构
- [x] PoW挖矿
- [x] 交易和UTXO
- [x] 钱包和签名
- [x] 持久化存储
- [x] CLI界面
- [ ] P2P网络
- [ ] 难度调整算法
- [ ] 脚本系统
- [ ] 轻节点支持

## 贡献

欢迎提交Issue和Pull Request！

1. Fork 项目
2. 创建分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 致谢

- [Bitcoin](https://bitcoin.org/) - 原始比特币协议
- [Mastering Bitcoin](https://github.com/bitcoinbook/bitcoinbook) - 学习资源

## 免责声明

这是一个**教育性质的演示项目**，用于学习区块链技术原理。它**不适合**用于生产环境或真实价值转移。

- 没有真正的P2P网络（使用文件模拟）
- 固定挖矿难度
- 简化的地址编码
- 明文存储私钥

---

Made with ❤️ for blockchain education
