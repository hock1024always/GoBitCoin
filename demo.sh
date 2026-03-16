#!/bin/bash

# Bitcoin Model Blockchain Demo Script
# 演示完整的区块链交易流程

echo "========================================"
echo "  Bitcoin Model Blockchain Demo"
echo "========================================"
echo ""

# 1. 创建矿工钱包
echo "Step 1: Creating miner wallet..."
./bin/miner -createwallet 2>/dev/null || echo "Miner wallet already exists"
echo ""

# 2. 初始化区块链
echo "Step 2: Initializing blockchain..."
./bin/miner -init 2>/dev/null || echo "Blockchain already initialized"
echo ""

# 3. 查看矿工余额
echo "Step 3: Checking miner balance..."
./bin/miner -balance
echo ""

# 4. 创建客户端钱包
echo "Step 4: Creating client wallet..."
./bin/client -createwallet 2>/dev/null || echo "Client wallet already exists"
echo ""

# 5. 列出钱包
echo "Step 5: Listing wallets..."
echo "Miner:"
./bin/miner -balance | grep "Address:"
echo ""
echo "Client:"
./bin/client -listwallets
echo ""

# 6. 矿工挖矿获得更多奖励
echo "Step 6: Mining a block to get more rewards..."
./bin/miner -mine -n=1
echo ""

# 7. 查看更新后的余额
echo "Step 7: Checking updated miner balance..."
./bin/miner -balance
echo ""

# 8. 显示区块链
echo "Step 8: Showing blockchain..."
./bin/miner -showchain
echo ""

echo "========================================"
echo "  Demo completed!"
echo "========================================"
echo ""
echo "Next steps:"
echo "1. Start miner in interactive mode: ./bin/miner -interactive"
echo "2. Start client in interactive mode: ./bin/client -interactive"
echo "3. Create transactions from client to miner"
echo "4. Mine blocks to confirm transactions"
