package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generate crea un JWT firmado para el panel de administración.
func Generate(secret, subject string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": subject,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Verify valida el token y devuelve el subject.
func Verify(secret, raw string) (string, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inválido")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("token inválido")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("token sin subject")
	}
	return sub, nil
}
