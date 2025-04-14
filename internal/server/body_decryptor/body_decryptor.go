package bodydecryptor

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Kopleman/metcol/internal/common/log"
)

func (bd *BodyDecryptor) DecryptBodyBytes(body []byte) ([]byte, error) {
	if bd.privateKey == nil {
		return nil, nil
	}
	decrypted, err := rsa.DecryptOAEP(sha256.New(), nil, bd.privateKey, body, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func (bd *BodyDecryptor) DecryptBody(body io.Reader) (io.Reader, error) {
	if bd.privateKey == nil {
		return body, nil
	}
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body bytes: %w", err)
	}

	decryptedBytes, decryptErr := bd.DecryptBodyBytes(bodyBytes)
	if decryptErr != nil {
		return nil, fmt.Errorf("failed to decrypt body bytes: %w", decryptErr)
	}

	return bytes.NewReader(decryptedBytes), nil
}

func (bd *BodyDecryptor) LoadPrivateKey(privateKeyPath string) error {
	if privateKeyPath == "" {
		return nil
	}

	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return errors.New("failed to parse private key PEM block")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	bd.privateKey = privKey

	return nil
}

type BodyDecryptor struct {
	logger     log.Logger
	privateKey *rsa.PrivateKey
}

func NewBodyDecryptor(logger log.Logger) *BodyDecryptor {
	bd := &BodyDecryptor{
		logger: logger,
	}

	return bd
}
