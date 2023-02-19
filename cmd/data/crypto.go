package data

import (
	"crypto/rand"
	"unsafe"
)

var alphabet = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func CreateNewObjectKey(strSize int) string {
	b := make([]byte, strSize)
	rand.Read(b)
	for i := 0; i < strSize; i++ {
		b[i] = alphabet[b[i]%byte(len(alphabet))]
	}
	return *(*string)(unsafe.Pointer(&b))
}
