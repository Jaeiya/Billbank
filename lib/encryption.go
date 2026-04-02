package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	timeCost   uint32 = 4
	memoryCost uint32 = 256 * 1024
	threads    uint8  = 4
	keyLen     uint32 = 32

	saltLen uint32 = 16
)

var ErrEncryptWithoutPassword = fmt.Errorf("tried to encrypt data without password")

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, keyLen)

	return base64.StdEncoding.EncodeToString(append(salt, hash...)), nil
}

func ValidatePassword(password string, storedPass string) (bool, error) {
	storedBytes, err := base64.StdEncoding.DecodeString(storedPass)
	if err != nil {
		return false, err
	}

	salt := storedBytes[:saltLen]
	hash := storedBytes[saltLen:]
	newHash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, keyLen)

	return subtle.ConstantTimeCompare(hash, newHash) == 1, nil
}

func EncryptData(data string, password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}
	key := deriveKey(password, salt)

	cBlock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(cBlock)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(data), nil)

	return base64.StdEncoding.EncodeToString(append(salt, cipherText...)), nil
}

/*
EncryptNonNil encrypts data using the password and returns the encrypted
string; however if data is nil it returns nil. Will only error if
password is nil.
*/
func EncryptNonNil(data *string, password *string) (any, error) /* nil|string */ {
	if data == nil {
		return nil, nil
	}
	if password == nil {
		return nil, ErrEncryptWithoutPassword
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

	salt := cipherText[:saltLen]
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
	nonce := cipherText[saltLen : int(saltLen)+nonceSize]

	cipherText = cipherText[int(saltLen)+nonceSize:]

	text, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, keyLen)
}
