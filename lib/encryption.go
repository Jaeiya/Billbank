package lib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
)

const (
	timeCost   uint32 = 2
	memoryCost uint32 = 128 * 1024
	threads    uint8  = 4
	saltLength uint32 = 16
	hashLength uint32 = 32
)

var ErrEncryptWithoutPassword = fmt.Errorf("tried to encrypt data without password")

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, hashLength)

	return base64.StdEncoding.EncodeToString(append(salt, hash...)), nil
}

func ValidatePassword(password string, storedPass string) (bool, error) {
	storedBytes, err := base64.StdEncoding.DecodeString(storedPass)
	if err != nil {
		return false, err
	}

	salt := storedBytes[:16]
	hash := storedBytes[16:]
	newHash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, hashLength)

	return bytes.Equal(hash, newHash), nil
}

func EncryptData(data string, password string) string {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}
	key := deriveKey(password, salt)

	cBlock, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	aesGCM, err := cipher.NewGCM(cBlock)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(data), nil)

	return base64.StdEncoding.EncodeToString(append(salt, cipherText...))
}

/*
EncryptNonNil encrypts data using the password and returns the encrypted
string; however if data is nil it returns nil.
*/
func EncryptNonNil(data *string, password *string) any /* nil|string */ {
	if data == nil {
		return nil
	}
	if password == nil {
		panic(ErrEncryptWithoutPassword)
	}
	return EncryptData(*data, *password)
}

func DecryptNonNil(data *string, password string) (*string, error) {
	if data == nil {
		return nil, nil
	}

	decData, err := DecryptData(*data, password)
	if err != nil {
		return nil, err
	}

	return &decData, err
}

func DecryptData(data string, password string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	salt := cipherText[:16]
	key := deriveKey(password, salt)

	cBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(cBlock)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce := cipherText[16 : 16+nonceSize]

	cipherText = cipherText[16+nonceSize:]

	text, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 100_000, 32, sha256.New)
}
