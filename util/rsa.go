package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lukmanlukmin/go-lib/log"
)

// ParsePubKeyFromString ...
func ParsePubKeyFromString(pubKey string) *rsa.PublicKey {
	pubKeyByte, err := hex.DecodeString(pubKey)
	if err != nil {
		log.Fatal(err)
	}
	pubAsli, err := x509.ParsePKIXPublicKey(pubKeyByte)
	if err != nil {
		log.Fatal(err)
	}
	return pubAsli.(*rsa.PublicKey)
}

// GenerateRSAKeyString ...
func GenerateRSAKeyString() (string, string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	pub := &priv.PublicKey

	privKey := x509.MarshalPKCS1PrivateKey(priv)
	pubKey, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(privKey), hex.EncodeToString(pubKey)
}

// DecodeHexRSAKeyString ...
func DecodeHexRSAKeyString(privateKey string, pubKeystring string) ([]byte, []byte) {
	privByte, err := hex.DecodeString(privateKey)
	if err != nil {
		log.Fatal(err)
	}
	pubByte, err := hex.DecodeString(pubKeystring)
	if err != nil {
		log.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privByte,
		}), pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubByte,
		})
}

// PubKeyToJWKKey ...
func PubKeyToJWKKey(pub *rsa.PublicKey, kid string) jwk.Key {
	key, err := jwk.New(pub)
	if err != nil {
		log.Fatal(err)
	}
	if kid != "" {
		if err := key.Set(jwk.KeyUsageKey, "sig"); err != nil {
			log.Fatal(fmt.Errorf("failed to set key usage: %w", err))
		}
		if err := key.Set(jwk.KeyIDKey, kid); err != nil {
			log.Fatal(fmt.Errorf("failed to set key usage: %w", err))
		}
	} else {
		if err := key.Set(jwk.KeyUsageKey, "sig"); err != nil {
			log.Fatal(fmt.Errorf("failed to set key usage: %w", err))
		}
	}
	if err := key.Set(jwk.AlgorithmKey, "RS256"); err != nil {
		log.Fatal(fmt.Errorf("failed to set key algorithm: %w", err))
	}
	return key
}
