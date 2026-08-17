package auth
// [J]SON [W]eb [T]oken


import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)
// Use the jwt.ParseWithClaims function to validate the signature of the JWT and extract the claims into a *jwt.Token struct. The keyFunc callback must return the same key type ([]byte) used when the token was signed. An error will be returned if the token is invalid or has expired.

// If all is well with the token, use the token.Claims interface to get access to the user's id from the claims (which should be stored in the Subject field). Return the id as a uuid.UUID.


func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	registerClaims := jwt.RegisteredClaims{}
	t, err := jwt.ParseWithClaims(
		tokenString, 
		&registerClaims,
		func(t *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		})
	if err != nil {
		return uuid.UUID{}, err
	}

	subject, err := t.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err

	}
	res, err := uuid.Parse(subject)
	if err != nil {
		return uuid.UUID{}, err
	}

	return res, nil
}





