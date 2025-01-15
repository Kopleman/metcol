package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/log"
)

type hashWriter struct {
	http.ResponseWriter
	key []byte
}

func newHashWriter(w http.ResponseWriter, key []byte) *hashWriter {
	return &hashWriter{
		ResponseWriter: w,
		key:            key,
	}
}

func (hw *hashWriter) Write(p []byte) (int, error) {
	n, err := hw.ResponseWriter.Write(p)
	if err != nil {
		return n, fmt.Errorf("hashMW: response write error: %w", err)
	}

	hash := hw.calcHashForBody(p)
	if hash != "" {
		hw.Header().Set(common.HashSHA256, hash)
	}
	return n, nil
}

func (hw *hashWriter) calcHashForBody(bodyBytes []byte) string {
	if len(hw.key) == 0 {
		return ""
	}
	if len(bodyBytes) == 0 {
		return ""
	}

	h := hmac.New(sha256.New, hw.key)
	h.Write(bodyBytes)
	hash := h.Sum(nil)
	hashString := hex.EncodeToString(hash)

	return hashString
}

func Hash(l log.Logger, keyString string) func(next http.Handler) http.Handler {
	hw := NewHashMiddleware(l, keyString)
	return hw.Handler
}

func NewHashMiddleware(l log.Logger, keyString string) *HashMiddleware {
	return &HashMiddleware{
		logger: l,
		key:    []byte(keyString),
	}
}

type HashMiddleware struct {
	logger log.Logger
	key    []byte
}

func (hw *HashMiddleware) validateHash(bodyBytes []byte, hashString string) (bool, error) {
	if hashString == "" {
		return true, nil
	}
	encrypted, err := hex.DecodeString(hashString)
	if err != nil {
		return false, fmt.Errorf("failed to decode string: %w", err)
	}
	h := hmac.New(sha256.New, hw.key)
	h.Write(bodyBytes)
	reqHash := h.Sum(nil)

	return hmac.Equal(reqHash, encrypted), nil
}

func (hw *HashMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(hw.key) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		hashW := newHashWriter(w, hw.key)

		hash := r.Header.Get(common.HashSHA256)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		valid, validateErr := hw.validateHash(bodyBytes, hash)
		if validateErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(hashW, r)
	})
}
