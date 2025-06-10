package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"

	"github.com/pkg/errors"
)

type DecryptRSA struct {
	key *rsa.PrivateKey
}

func NewDecryptRSA(privateKeyFile string) (*DecryptRSA, error) {
	key, err := loadPrivateKey(privateKeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "错误加载私钥")
	}
	return &DecryptRSA{key: key}, nil
}

func (d *DecryptRSA) Decrypt(encryptedData string) (string, error) {
	return decrypt(d.key, encryptedData)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "错误读取私钥文件")
	}
	block, _ := pem.Decode(file)
	if block == nil {
		return nil, errors.Wrap(err, "错误解码PEM块")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "错误解析私钥")
	}
	return key, nil
}

func decrypt(key *rsa.PrivateKey, encryptedData string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", errors.Wrap(err, "Base64解码失败")
	}

	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		key,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", errors.Wrap(err, "解密失败")
	}

	return string(plaintext), nil
}
