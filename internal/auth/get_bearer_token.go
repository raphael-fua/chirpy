package auth


import (
	// "github.com/golang-jwt/jwt/v5"
	"errors"
	"net/http"
	"strings"
)


func GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", errors.New("authorization header is empty") 
	}
	words := strings.Fields(header) 
	if len(words) < 2 {
		return "", errors.New("authorization header should be at least two words")
	}
	if words[0] != "Bearer" {
		return "", errors.New("token string should be preceded by BEARER")
	}
	tokenString := words[1]
	return tokenString, nil
}
