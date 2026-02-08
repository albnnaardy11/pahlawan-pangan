// Package main provides a utility to generate RSA key pairs for JWT RS256 signing.
// This tool generates both private and public keys with secure file permissions.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func main() {
	// 1. Generate Private Key (2048 bits for Bank Grade)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// 2. Encode and Save Private Key
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	if err := os.WriteFile("private.pem", privPEM, 0600); err != nil {
		panic(err)
	}

	// 3. Extract and Save Public Key
	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})
	// Public keys can be world-readable, but we use 0600 for consistency with security best practices
	if err := os.WriteFile("public.pem", pubPEM, 0600); err != nil {
		panic(err)
	}

	println("✅ RSA Keys (RS256) Generated: private.pem & public.pem")
}
