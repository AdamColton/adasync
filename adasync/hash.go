package adasync

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
)

type Hash [md5.Size]byte

func HashFromBytes(bs []byte) *Hash {
	if len(bs) != md5.Size {
		panic("Requires bytes equals md5.Size")
	}
	hash := Hash{}
	copy(hash[:], bs)
	return &hash
}

func (hash *Hash) String() string {
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (a *Hash) Equal(b *Hash) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return bytes.Equal(a[:], b[:])
}
