package auth


import (
	"crypto/rand"
	"encoding/hex"
)


func MakeRefreshToken() string {
	token := make([]byte, 32)
	_, _  = rand.Read(token)
	return hex.EncodeToString(token)
}


