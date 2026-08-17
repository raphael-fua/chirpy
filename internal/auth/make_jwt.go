// [J]SON [W]eb [T]oken
package auth


import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)



func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	t1 := time.Now().UTC()
	t2 := t1.Add(expiresIn)

	now := jwt.NewNumericDate(t1)
	then := jwt.NewNumericDate(t2)

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: now,
		ExpiresAt: then,
		Subject: userID.String(),
	})

	res, err := t.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return res, nil
}




