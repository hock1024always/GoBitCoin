package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
)

const (
	walletFile = "wallet.dat"
	walletDir  = "./wallets"
)

// Wallet 钱包结构
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
	Address    string
}

// NewWallet 创建新钱包
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	address := GenerateAddress(publicKey)

	wallet := &Wallet{
		PrivateKey: *privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}

	return wallet, nil
}

// GenerateAddress 从公钥生成地址
func GenerateAddress(pubKey []byte) string {
	// 1. SHA256哈希
	pubHash := sha256.Sum256(pubKey)

	// 2. RIPEMD160哈希（简化版，直接用SHA256的前20字节）
	versionedHash := append([]byte{0x00}, pubHash[:20]...)

	// 3. 双重SHA256校验和
	checksum := checksum(versionedHash)

	// 4. 添加校验和
	fullHash := append(versionedHash, checksum...)

	// 5. Base58编码（简化版，用hex编码）
	return hex.EncodeToString(fullHash)
}

// checksum 计算校验和
func checksum(payload []byte) []byte {
	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])
	return hash2[:4]
}

// HashPubKey 公钥哈希（用于锁定脚本）
func HashPubKey(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)
	return pubHash[:20]
}

// GetPubKeyHashFromAddress 从地址提取PubKeyHash
// 简化版：地址就是PubKeyHash的hex编码（去掉版本和校验和）
func GetPubKeyHashFromAddress(address string) []byte {
	// 地址格式: version(1 byte) + PubKeyHash(20 bytes) + checksum(4 bytes)
	// 总共25字节，hex编码后50字符
	// 简化处理：直接解码hex，取中间20字节
	if len(address) < 40 {
		return []byte(address)
	}
	// 解码hex
	decoded, err := hex.DecodeString(address)
	if err != nil || len(decoded) < 21 {
		return HashPubKey([]byte(address))
	}
	// 返回PubKeyHash部分（跳过版本字节）
	return decoded[1:21]
}

// GetPrivateKeyBytes 获取私钥字节
func (w *Wallet) GetPrivateKeyBytes() []byte {
	privKeyBytes, _ := x509.MarshalECPrivateKey(&w.PrivateKey)
	return privKeyBytes
}

// SaveToFile 保存钱包到文件
func (w *Wallet) SaveToFile(nodeID string) error {
	filename := getWalletFile(nodeID)

	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	walletData := struct {
		PrivateKey []byte `json:"private_key"`
		PublicKey  []byte `json:"public_key"`
		Address    string `json:"address"`
	}{
		PrivateKey: w.GetPrivateKeyBytes(),
		PublicKey:  w.PublicKey,
		Address:    w.Address,
	}

	data, err := json.MarshalIndent(walletData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadWalletFromFile 从文件加载钱包
func LoadWalletFromFile(nodeID string) (*Wallet, error) {
	filename := getWalletFile(nodeID)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var walletData struct {
		PrivateKey []byte `json:"private_key"`
		PublicKey  []byte `json:"public_key"`
		Address    string `json:"address"`
	}

	if err := json.Unmarshal(data, &walletData); err != nil {
		return nil, err
	}

	privateKey, err := x509.ParseECPrivateKey(walletData.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey: *privateKey,
		PublicKey:  walletData.PublicKey,
		Address:    walletData.Address,
	}, nil
}

// WalletExists 检查钱包是否存在
func WalletExists(nodeID string) bool {
	filename := getWalletFile(nodeID)
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// getWalletFile 获取钱包文件路径
func getWalletFile(nodeID string) string {
	if nodeID == "" {
		return filepath.Join(walletDir, walletFile)
	}
	return filepath.Join(walletDir, fmt.Sprintf("wallet_%s.dat", nodeID))
}

// SignData 使用私钥签名数据
func SignData(privKey []byte, data []byte) ([]byte, error) {
	privateKey, err := x509.ParseECPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, data)
	if err != nil {
		return nil, err
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature 验证签名
func VerifySignature(pubKey []byte, data []byte, signature []byte) bool {
	if len(signature) != 64 {
		return false
	}

	// 解析公钥
	curve := elliptic.P256()
	x := new(big.Int).SetBytes(pubKey[:len(pubKey)/2])
	y := new(big.Int).SetBytes(pubKey[len(pubKey)/2:])

	publicKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	// 解析签名
	r := new(big.Int).SetBytes(signature[:len(signature)/2])
	s := new(big.Int).SetBytes(signature[len(signature)/2:])

	return ecdsa.Verify(publicKey, data, r, s)
}

// GetPubKeyFromAddress 从地址获取公钥（简化版，实际应该通过地址反查）
func GetPubKeyFromAddress(address string) ([]byte, error) {
	// 这里简化处理，实际应该维护地址到公钥的映射
	// 在真实场景中，地址是从公钥派生出来的，不能反向推导
	return nil, fmt.Errorf("cannot derive public key from address")
}
