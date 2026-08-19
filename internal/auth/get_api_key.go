package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", errors.New("authorization header is empty")
	}
	words := strings.Fields(header)
	if len(words) < 2 {
		return "", errors.New("authorization header should be at least two words")
	}
	if words[0] != "ApiKey" {
		return "", errors.New("first word of authorization header should be 'apiKey'")
	}
	tokenString := words[1]
	return tokenString, nil
}


