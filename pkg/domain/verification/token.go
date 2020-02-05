package verification

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const KeyLength = 32

func GenerateDomainKey() string {
	bytes := make([]byte, KeyLength)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func GenerateDomainToken(key string, nonce string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(nonce))
	mac := h.Sum(nil)
	return hex.EncodeToString(mac)
}
