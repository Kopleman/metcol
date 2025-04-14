// Package cryptokeysgenerator used to create keys pair.
package cryptokeysgenerator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// Generator instance.
type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateKeys(privateKeyPath, publicKeyPath string) error {
	if publicKeyPath == "" || privateKeyPath == "" {
		return errors.New("publicKeyPath or privateKeyPath is empty")
	}

	if publicKeyPath == privateKeyPath {
		return errors.New("publicKeyPath is equal to privateKeyPath")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key pair: %w", err)
	}

	if pubExportErr := g.exportPublicKeyToFile(&privateKey.PublicKey, publicKeyPath); pubExportErr != nil {
		return fmt.Errorf("failed to export public key to file: %w", pubExportErr)
	}

	if privateExportErr := g.exportPrivateKeyToFile(privateKey, privateKeyPath); privateExportErr != nil {
		return fmt.Errorf("failed to export private key to file: %w", privateExportErr)
	}

	return nil
}

func (g *Generator) exportPublicKeyToFile(publicKey *rsa.PublicKey, pathToFile string) error {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to serialize public key: %w", err)
	}
	pubKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubKeyBytes,
		},
	)

	file, fileErr := g.getFileDescriptor(pathToFile)
	if fileErr != nil {
		return fmt.Errorf("failed to get file descriptor for pub-key: %w", fileErr)
	}

	if _, writeErr := file.Write(pubKeyPEM); writeErr != nil {
		return fmt.Errorf("failed to write public key to file: %w", writeErr)
	}

	return nil
}

func (g *Generator) exportPrivateKeyToFile(privateKey *rsa.PrivateKey, pathToFile string) error {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)

	file, fileErr := g.getFileDescriptor(pathToFile)
	if fileErr != nil {
		return fmt.Errorf("failed to get file descriptor for private-key: %w", fileErr)
	}

	if _, writeErr := file.Write(privKeyPEM); writeErr != nil {
		return fmt.Errorf("failed to write private key to file: %w", writeErr)
	}

	return nil
}

func (g *Generator) getFileDescriptor(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666) //nolint:all // different lint behavior on perm var
	if err != nil {
		return nil, fmt.Errorf("could not access file by path '%s': %w", path, err)
	}

	if err = file.Truncate(0); err != nil {
		return nil, fmt.Errorf("could not file: %w", err)
	}

	return file, nil
}
