package bodydecryptor

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
	"testing"
	"testing/iotest"

	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPrivateKey_EmptyPath(t *testing.T) {
	bd := NewBodyDecryptor(log.MockLogger{})
	err := bd.LoadPrivateKey("")
	assert.NoError(t, err)
	assert.Nil(t, bd.privateKey)
}

func TestLoadPrivateKey_FileNotExists(t *testing.T) {
	bd := NewBodyDecryptor(log.MockLogger{})
	err := bd.LoadPrivateKey("nonexistent.pem")
	assert.ErrorContains(t, err, "unable to read private key file")
}

func TestLoadPrivateKey_InvalidPEM(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testkey")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte("invalid pem data"))
	require.NoError(t, err)
	tmpfile.Close()

	bd := NewBodyDecryptor(log.MockLogger{})
	err = bd.LoadPrivateKey(tmpfile.Name())
	assert.ErrorContains(t, err, "failed to parse private key PEM block")
}

func TestLoadPrivateKey_ValidKey(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}
	pemData := pem.EncodeToMemory(block)

	tmpfile, err := os.CreateTemp("", "testkey")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(pemData)
	require.NoError(t, err)
	tmpfile.Close()

	bd := NewBodyDecryptor(log.MockLogger{})
	err = bd.LoadPrivateKey(tmpfile.Name())
	assert.NoError(t, err)
	require.NotNil(t, bd.privateKey)
	assert.Equal(t, privKey.D, bd.privateKey.D)
}

func TestDecryptBodyBytes_NoPrivateKey(t *testing.T) {
	bd := NewBodyDecryptor(log.MockLogger{})
	decrypted, err := bd.DecryptBodyBytes([]byte("test"))
	assert.NoError(t, err)
	assert.Nil(t, decrypted)
}

func TestDecryptBodyBytes_Success(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	plaintext := []byte("hello world")
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, plaintext, nil)
	require.NoError(t, err)

	bd := NewBodyDecryptor(log.MockLogger{})
	bd.privateKey = privKey

	decrypted, err := bd.DecryptBodyBytes(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptBodyBytes_InvalidCiphertext(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	bd := NewBodyDecryptor(log.MockLogger{})
	bd.privateKey = privKey

	encrypted := []byte("invalid ciphertext")
	decrypted, err := bd.DecryptBodyBytes(encrypted)
	assert.Error(t, err)
	assert.Nil(t, decrypted)
}

func TestDecryptBody_NoPrivateKey(t *testing.T) {
	originalData := []byte("test data")
	reader := bytes.NewReader(originalData)

	bd := NewBodyDecryptor(log.MockLogger{})
	resultReader, err := bd.DecryptBody(reader)
	assert.NoError(t, err)

	resultData, err := io.ReadAll(resultReader)
	assert.NoError(t, err)
	assert.Equal(t, originalData, resultData)
}

func TestDecryptBody_Success(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	plaintext := []byte("hello world reader")
	encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, plaintext, nil)
	require.NoError(t, err)

	bd := NewBodyDecryptor(log.MockLogger{})
	bd.privateKey = privKey

	encryptedReader := bytes.NewReader(encrypted)
	resultReader, err := bd.DecryptBody(encryptedReader)
	assert.NoError(t, err)

	decryptedData, err := io.ReadAll(resultReader)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decryptedData)
}

func TestDecryptBody_ReadError(t *testing.T) {
	errReader := iotest.ErrReader(errors.New("read error"))
	bd := NewBodyDecryptor(log.MockLogger{})
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	bd.privateKey = privKey

	_, err = bd.DecryptBody(errReader)
	assert.ErrorContains(t, err, "failed to read body bytes")
}
