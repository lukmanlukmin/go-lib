package util

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
)

// GenerateJWT generate jwt
func GenerateJWT(secret string, duration time.Duration, data map[string]interface{}) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(duration).Unix(),
	}
	for key, value := range data {
		claims[key] = value
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT validate jwt
func ValidateJWT(secret, tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// pastikan algoritma sesuai
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Convert to map[string]interface{}
		result := make(map[string]interface{})
		for k, v := range claims {
			result[k] = v
		}
		return result, nil
	}

	return nil, fmt.Errorf("invalid or expired token")
}

// GenerateJWTRSA generates a JWT signed with RSA private key
func GenerateJWTRSA(privateKey *rsa.PrivateKey, kid string, duration time.Duration, data map[string]interface{}) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(duration).Unix(),
		"iat": time.Now().Unix(),
	}

	for key, value := range data {
		claims[key] = value
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set kid in header
	token.Header["kid"] = kid

	// Sign with RSA private key
	return token.SignedString(privateKey)
}

// ValidateJWTRSA validates JWT using JWKS from URL
func ValidateJWTRSA(jwksURL, tokenString string) (map[string]interface{}, error) {
	// Fetch JWKS
	set, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Parse token and verify using key set
	keyfunc := func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		key, found := set.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("key with kid %s not found in JWKS", kid)
		}

		var pubkey rsa.PublicKey
		if err := key.Raw(&pubkey); err != nil {
			return nil, fmt.Errorf("invalid RSA public key: %w", err)
		}
		return &pubkey, nil
	}

	token, err := jwt.Parse(tokenString, keyfunc)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result := make(map[string]interface{})
		for k, v := range claims {
			result[k] = v
		}
		return result, nil
	}

	return nil, fmt.Errorf("invalid or expired token")
}
