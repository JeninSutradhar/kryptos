/// Package crypto provides cryptographic utilities for secure data encryption and decryption.
// This package offers functions to:
// - Derive encryption keys from passwords using the scrypt key derivation function.
// - Generate random salts for use in key derivation.
// - Encrypt and decrypt data using AES-256-GCM (Authenticated Encryption with Associated Data).

package crypto

import (
	"crypto/aes" 
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptN       = 16384  // scrypt parameter N (controls the cost of the key derivation)
	scryptR       = 8      // scrypt parameter R (controls memory usage)
	scryptP       = 1      // scrypt parameter P (parallelization factor)
	scryptSaltLen = 16     // Length of the salt in bytes (random value used in key derivation)
)

// DeriveKeyFromPassword derives an encryption key from a master password using scrypt.
// It takes a master password and a salt as inputs and returns a 32-byte encryption key
// suitable for use with AES-256 encryption. The function returns an error if the key
// derivation fails.
func DeriveKeyFromPassword(masterPassword string, salt []byte) ([]byte, error) {
	dk, err := scrypt.Key([]byte(masterPassword), salt, scryptN, scryptR, scryptP, 32) // 32-byte key for AES-256
	if err != nil {
		return nil, err
	}
	return dk, nil
}

// GenerateSalt generates a random salt of 16 bytes.
// It returns the generated salt and an error if the random data generation fails.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, scryptSaltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// EncryptData encrypts data using AES-256-GCM (Authenticated Encryption with Associated Data).
// It takes the plaintext data and a 32-byte encryption key as inputs, generates a random nonce,
// encrypts the data, and returns the base64-encoded ciphertext. The function returns an error
// if any step of the encryption process fails.
func EncryptData(data []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil // Base64 encode for text storage
}

// DecryptData decrypts data encrypted with AES-256-GCM.
// It takes the base64-encoded ciphertext and a 32-byte encryption key as inputs,
// extracts the nonce, decrypts the data, and returns the plaintext. The function
// returns an error if any step of the decryption process fails.
func DecryptData(encryptedData string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
