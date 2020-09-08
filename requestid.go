package contracts

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

type ctxKey int

const (
	RequestIDKey ctxKey = iota + 1
)

func shortID() string {
	b := make([]byte, 10)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}
