package collection

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
)

type Hash [md5.Size]byte

func HashFromBytes(bs []byte) *Hash {
	if len(bs) != 16 {
		panic("Requires 16 bytes")
	}
	hash := &Hash{}
	for i, b := range bs {
		hash[i] = b
	}
	return hash
}

func (hash *Hash) String() string {
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (a *Hash) Equal(b *Hash) bool {
	if a == nil && b == nil {
		return true
	}
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}
	return bytes.Equal(a[:], b[:])
}
