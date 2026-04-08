package internal

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryption(t *testing.T) {
	t.Parallel()
	passEncTests := []struct {
		name     string
		password string
	}{
		{"encrypt empty password", ""},
		{"encrypt simple password", "password123"},
		{"encrypt complex password", "P@$$w0rd!123~`[]{}"},
		{"encrypt long password", strings.Repeat("abcdefghijklmnopqrstuvwxyz", 10)},
		{"encrypt unicode password", "пароль密碼パスワード비밀번호"},
	}

	for _, tt := range passEncTests {
		t.Run("should "+tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)
			hash1, err := HashPassword(tt.password)
			r.NoError(err)

			a.NotEmpty(hash1, "expected to contain hash")

			hash2, err := HashPassword(tt.password)
			r.NoError(err)
			a.NotEqual(hash1, hash2, "expected same password to produce different hash")

			decoded, err := base64.StdEncoding.DecodeString(hash1)
			a.NoError(err, "expected a valid base64 encoding")

			expectedLen := saltLen + keyLen
			a.Len(decoded, int(expectedLen), "expected decoded length to be saltLen + hashLen")
		})
	}

	passwordTests := []struct {
		name        string
		password    string
		changeInput bool
	}{
		{"validate empty password", "", false},
		{"validate simple password", "password123", false},
		{"validate complex password", "P@$$w0rd!123~`[]{}}", false},
		{"fail to validate invalid pass", "password123", true},
	}

	for _, tt := range passwordTests {
		t.Run("should "+tt.name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)
			a := assert.New(t)

			hash, err := HashPassword(tt.password)
			r.NoError(err)

			validPass := tt.password
			if tt.changeInput {
				validPass = "wrong" + validPass
			}

			valid, err := ValidatePassword(validPass, hash)
			r.NoError(err)

			if tt.changeInput {
				a.False(valid, "expected password to be incorrect")
			} else {
				a.True(valid, "expected password to be correct")
			}
		})
	}

	// Test with invalid hash format
	t.Run("should fail pass validation on invalid param", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		_, err := ValidatePassword("password", "not-base64")
		a.Error(err, "expected invalid base64")
	})

	dataEncTests := []struct {
		name     string
		data     string
		password string
		wantErr  bool
	}{
		{"encrypt empty data, empty password", "", "", false},
		{"encrypt data, with password", "", "password123", false},
		{"encrypt data, empty password", "secret data", "", false},
		{"encrypt data and password", "secret data", "password123", false},
		{
			"encrypt data",
			strings.Repeat("abcdefghijklmnopqrstuvwxyz", 100),
			"password123",
			false,
		},
		{"encrypt data", "секретные данные機密數據秘密データ비밀 데이터", "password123", false},
	}

	for _, tt := range dataEncTests {
		t.Run("should "+tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)
			// Test encryption
			encrypted, err := EncryptData(tt.data, tt.password)
			if !tt.wantErr {
				r.NoError(err)
			}

			if tt.wantErr {
				a.Error(err, "expected to get error")
			}

			encryptedBytes, err := base64.StdEncoding.DecodeString(encrypted)
			r.NoError(err, "expected base64 encoded encryption")
			a.NotEmpty(encrypted, "expected encrypted data")
			a.NotEqual(tt.data, encrypted, "expected data to be encrypted")

			encrypted2, err := EncryptData(tt.data, tt.password)
			r.NoError(err)
			a.NotEqual(encrypted, encrypted2, "expected different output with same input")

			decrypted, err := DecryptData(encrypted, tt.password)
			r.NoError(err, "expect successful decryption operation")
			a.Equal(tt.data, decrypted, "expected data to be decrypted")

			a.GreaterOrEqual(
				len(encryptedBytes),
				16*2,
				"expected saltLen (16 bytes) + minimum GCM overhead",
			)
		})
	}
}

func TestEncryptNonNil(t *testing.T) {
	t.Parallel()
	encNonNilTests := []struct {
		should      string
		data        *string
		password    *string
		expectedErr error
		expectNil   bool
	}{
		{
			should:    "return nil with nil data",
			data:      nil,
			password:  new("password"),
			expectNil: true,
		},
		{
			should:      "panic with nil password",
			data:        new("secret"),
			password:    nil,
			expectedErr: ErrEncryptWithoutPassword,
		},
		{
			should:    "encode non nil values",
			data:      new("secret"),
			password:  new("password"),
			expectNil: false,
		},
	}

	for _, tt := range encNonNilTests {
		t.Run("should "+tt.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)

			result, err := EncryptNonNil(tt.data, tt.password)
			if tt.expectedErr == nil {
				r.NoError(err)
			} else {
				a.Error(ErrEncryptWithoutPassword, "expected error with nil password")
				return
			}

			if tt.expectNil {
				a.Nil(result, "expected result to be nil because data is nil")
				return
			}

			a.NotNil(result, "expected a result")

			resultStr, isType := result.(string)
			a.True(isType, "expected result to be a string")
			a.NotEmpty(resultStr, "expected result to have a value")

			// Verify we can decrypt the result
			if tt.data != nil && tt.password != nil {
				decrypted, err := DecryptData(resultStr, *tt.password)
				a.NoError(err, "expected to decrypt")
				a.Equal(*tt.data, decrypted, "expected to get original data")
			}
		})
	}
}

func TestDecryptNonNil(t *testing.T) {
	t.Parallel()
	password := "password123"
	plaintext := "secret data"
	encrypted, _ := EncryptData(plaintext, password)

	tests := []struct {
		should   string
		data     *string
		password string
		wantErr  bool
		wantNil  bool
	}{
		{
			should:   "return nil with nil data",
			data:     nil,
			password: password,
			wantErr:  false,
			wantNil:  true,
		},
		{
			should:   "decrypt valid data and password",
			data:     &encrypted,
			password: password,
			wantErr:  false,
			wantNil:  false,
		},
		{
			should:   "catch wrong password",
			data:     &encrypted,
			password: "wrong",
			wantErr:  true,
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run("should "+tt.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			r := require.New(t)
			result, err := DecryptNonNil(tt.data, tt.password)
			if !tt.wantErr {
				r.NoError(err, "expected successful operation")
			} else {
				a.ErrorContains(err, "authentication failed", "expected failed authentication error")
			}

			if tt.wantNil {
				a.Nil(result, "expected nil value with nil data")
				return
			}

			if !tt.wantErr {
				a.NotNil(result, "expected to get a returned value")
				if result != nil {
					a.Equal(plaintext, *result, "expected to get decrypted data")
				}
			}
		})
	}
}

func TestDeriveKey(t *testing.T) {
	t.Parallel()
	// Test that the same password and salt produce the same key
	password := "password123"
	salt := bytes.Repeat([]byte{1}, int(saltLen))

	key1 := deriveKey(password, salt)
	key2 := deriveKey(password, salt)

	a := assert.New(t)

	a.Equal(key1, key2, "expected same input to produce same output")

	// Test that different passwords produce different keys
	key3 := deriveKey("different"+password, salt)
	a.NotEqual(key1, key3, "expected different passwords to produce different keys")

	// Test that different salts produce different keys
	salt2 := bytes.Repeat([]byte{2}, int(saltLen))
	key4 := deriveKey(password, salt2)
	a.NotEqual(key1, key4, "expected different salts to produce different keys")

	// Test key length
	a.Len(key1, int(keyLen), "expected key length to be 32")
}
