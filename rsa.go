package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// 默认密钥长度
	DefaultKeySize = 2048

	// 默认密钥存储目录
	DefaultKeyDir = "keys"

	// 默认私钥文件名
	DefaultPrivateKeyFile = "private.pem"

	// 默认公钥文件名
	DefaultPublicKeyFile = "public.pem"
)

// RSAKeyManager 管理RSA密钥的生成、保存和加载
type RSAKeyManager struct {
	keyDir         string
	privateKeyFile string
	publicKeyFile  string
}

// NewRSAKeyManager 创建新的密钥管理器
func NewRSAKeyManager() *RSAKeyManager {
	return &RSAKeyManager{
		keyDir:         DefaultKeyDir,
		privateKeyFile: DefaultPrivateKeyFile,
		publicKeyFile:  DefaultPublicKeyFile,
	}
}

// WithKeyDir 设置密钥存储目录
func (r *RSAKeyManager) WithKeyDir(dir string) *RSAKeyManager {
	r.keyDir = dir
	return r
}

// WithKeyFiles 设置密钥文件名
func (r *RSAKeyManager) WithKeyFiles(privateFile, publicFile string) *RSAKeyManager {
	r.privateKeyFile = privateFile
	r.publicKeyFile = publicFile
	return r
}

// Init 初始化密钥系统
func (r *RSAKeyManager) Init() error {
	// 创建密钥目录
	if err := os.MkdirAll(r.keyDir, 0700); err != nil {
		return fmt.Errorf("无法创建密钥目录: %w", err)
	}

	// 确保密钥存在
	_, err := r.ensureKeys()
	return err
}

// GenerateKeyPair 生成新的RSA密钥对
func (r *RSAKeyManager) GenerateKeyPair(keySize int) error {
	if keySize < 512 {
		keySize = DefaultKeySize
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return fmt.Errorf("密钥生成失败: %w", err)
	}

	// 保存私钥
	if err := r.SavePrivateKey(privateKey); err != nil {
		return err
	}

	// 保存公钥
	return r.SavePublicKey(&privateKey.PublicKey)
}

// SavePrivateKey 保存私钥到文件
func (r *RSAKeyManager) SavePrivateKey(key *rsa.PrivateKey) error {
	path := filepath.Join(r.keyDir, r.privateKeyFile)

	// 使用PKCS#8格式编码私钥
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("PKCS#8编码失败: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	if err := pem.Encode(file, block); err != nil {
		return fmt.Errorf("PEM编码失败: %w", err)
	}

	return nil
}

// SavePublicKey 保存公钥到文件
func (r *RSAKeyManager) SavePublicKey(key *rsa.PublicKey) error {
	path := filepath.Join(r.keyDir, r.publicKeyFile)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return fmt.Errorf("公钥序列化失败: %w", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	return pem.Encode(file, publicKeyPEM)
}

// LoadPrivateKey 从文件加载私钥
func (r *RSAKeyManager) LoadPrivateKey() (*rsa.PrivateKey, error) {
	path := filepath.Join(r.keyDir, r.privateKeyFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return ParsePrivateKeyPEM(string(data))
}

// LoadPublicKey 从文件加载公钥
func (r *RSAKeyManager) LoadPublicKey() (*rsa.PublicKey, error) {
	path := filepath.Join(r.keyDir, r.publicKeyFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return ParsePublicKeyPEM(string(data))
}

// PrivateKeyPath 获取私钥路径
func (r *RSAKeyManager) PrivateKeyPath() string {
	return filepath.Join(r.keyDir, r.privateKeyFile)
}

// PublicKeyPath 获取公钥路径
func (r *RSAKeyManager) PublicKeyPath() string {
	return filepath.Join(r.keyDir, r.publicKeyFile)
}

// 确保密钥存在（内部使用）
func (r *RSAKeyManager) ensureKeys() (*rsa.PrivateKey, error) {
	privPath := r.PrivateKeyPath()

	if _, err := os.Stat(privPath); err == nil {
		return r.LoadPrivateKey()
	}

	// 生成新密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, DefaultKeySize)
	if err != nil {
		return nil, fmt.Errorf("密钥生成失败: %w", err)
	}

	// 保存私钥
	if err := r.SavePrivateKey(privateKey); err != nil {
		return nil, err
	}

	// 保存公钥
	if err := r.SavePublicKey(&privateKey.PublicKey); err != nil {
		return nil, err
	}

	return privateKey, nil
}

// RSAEncryptor RSA加密器
type RSAEncryptor struct {
	publicKey *rsa.PublicKey
}

// NewRSAEncryptorFromKey 从公钥对象创建加密器
func NewRSAEncryptorFromKey(publicKey *rsa.PublicKey) *RSAEncryptor {
	return &RSAEncryptor{publicKey: publicKey}
}

// NewRSAEncryptorFromFile 从公钥文件创建加密器
func NewRSAEncryptorFromFile(publicKeyPath string) (*RSAEncryptor, error) {
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败: %w", err)
	}
	return NewRSAEncryptorFromPEM(string(data))
}

// NewRSAEncryptorFromPEM 从公钥PEM文本创建加密器
func NewRSAEncryptorFromPEM(publicKeyPEM string) (*RSAEncryptor, error) {
	publicKey, err := ParsePublicKeyPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}
	return &RSAEncryptor{publicKey: publicKey}, nil
}

// Encrypt 加密文本
func (e *RSAEncryptor) Encrypt(text string) (string, error) {
	// 检查文本长度
	maxLen := e.publicKey.Size() - 2*sha256.New().Size() - 2
	if len(text) > maxLen {
		return "", fmt.Errorf("文本过长(最大%d字符)，请缩短内容", maxLen)
	}

	// 加密
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		e.publicKey,
		[]byte(text),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("加密失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSADecryptor RSA解密器
type RSADecryptor struct {
	privateKey *rsa.PrivateKey
}

// NewRSADecryptorFromKey 从私钥对象创建解密器
func NewRSADecryptorFromKey(privateKey *rsa.PrivateKey) *RSADecryptor {
	return &RSADecryptor{privateKey: privateKey}
}

// NewRSADecryptorFromFile 从私钥文件创建解密器
func NewRSADecryptorFromFile(privateKeyPath string) (*RSADecryptor, error) {
	data, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}
	return NewRSADecryptorFromPEM(string(data))
}

// NewRSADecryptorFromPEM 从私钥PEM文本创建解密器
func NewRSADecryptorFromPEM(privateKeyPEM string) (*RSADecryptor, error) {
	privateKey, err := ParsePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	return &RSADecryptor{privateKey: privateKey}, nil
}

// Decrypt 解密文本
func (d *RSADecryptor) Decrypt(encryptedText string) (string, error) {
	// 解码Base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 解密
	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		d.privateKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}

// ParsePrivateKeyPEM 从PEM文本解析私钥
func ParsePrivateKeyPEM(pemText string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("无效的PEM格式")
	}

	// 尝试解析为PKCS#8格式
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("不是RSA私钥")
		}
	}

	// 如果PKCS#8解析失败，尝试PKCS#1格式
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, errors.New("无法解析私钥格式")
}

// ParsePublicKeyPEM 从PEM文本解析公钥
func ParsePublicKeyPEM(pemText string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("无效的PEM格式")
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("不是公钥: %s", block.Type)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("不是RSA公钥")
	}

	return rsaPub, nil
}
