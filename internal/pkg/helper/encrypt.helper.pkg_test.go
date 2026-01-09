package helper

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDips(t *testing.T) {
	// Set required environment variables for testing
	originalPassword := os.Getenv("DIPS_PASSWORD")
	originalMethod := os.Getenv("DIPS_AES_METHOD")
	originalIVLength := os.Getenv("DIPS_IV_LENGTH")

	defer func() {
		os.Setenv("DIPS_PASSWORD", originalPassword)
		os.Setenv("DIPS_AES_METHOD", originalMethod)
		os.Setenv("DIPS_IV_LENGTH", originalIVLength)
	}()

	// Set valid test environment
	os.Setenv("DIPS_PASSWORD", "12345678901234567890123456789012") // 32 bytes
	os.Setenv("DIPS_AES_METHOD", "aes-256-cbc")
	os.Setenv("DIPS_IV_LENGTH", "16")

	t.Run("successful encryption", func(t *testing.T) {
		text := "test message"
		encrypted, err := EncryptDips(text)

		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.Contains(t, encrypted, ":")

		parts := strings.Split(encrypted, ":")
		assert.Len(t, parts, 2)
		assert.NotEmpty(t, parts[0]) // IV
		assert.NotEmpty(t, parts[1]) // encrypted data
	})

	t.Run("empty text", func(t *testing.T) {
		encrypted, err := EncryptDips("")

		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.Contains(t, encrypted, ":")
	})

	t.Run("invalid cipher key length", func(t *testing.T) {
		os.Setenv("DIPS_PASSWORD", "short") // invalid length

		_, err := EncryptDips("test")
		assert.Error(t, err)
	})

	t.Run("unsupported method", func(t *testing.T) {
		os.Setenv("DIPS_PASSWORD", "12345678901234567890123456789012")
		os.Setenv("DIPS_AES_METHOD", "aes-128-cbc") // unsupported

		_, err := EncryptDips("test")
		assert.Error(t, err)
	})

	t.Run("custom IV length", func(t *testing.T) {
		os.Setenv("DIPS_PASSWORD", "12345678901234567890123456789012")
		os.Setenv("DIPS_AES_METHOD", "aes-256-cbc")
		os.Setenv("DIPS_IV_LENGTH", "16") // Must be 16 for AES block size

		encrypted, err := EncryptDips("test")
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)
	})
}

func TestDecryptDips(t *testing.T) {
	// Set required environment variables for testing
	originalPassword := os.Getenv("DIPS_PASSWORD")
	defer func() {
		os.Setenv("DIPS_PASSWORD", originalPassword)
	}()

	os.Setenv("DIPS_PASSWORD", "12345678901234567890123456789012") // 32 bytes

	t.Run("successful decryption", func(t *testing.T) {
		os.Setenv("DIPS_AES_METHOD", "aes-256-cbc")

		// First encrypt a test message
		testMessage := "test decryption message"
		encrypted, err := EncryptDips(testMessage)
		assert.NoError(t, err)

		// Then decrypt it
		decrypted, err := DecryptDips(encrypted)
		assert.NoError(t, err)
		assert.Equal(t, testMessage, decrypted)
	})

	t.Run("invalid format - no colon", func(t *testing.T) {
		_, err := DecryptDips("invalidformat")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input text")
	})

	t.Run("invalid format - empty parts", func(t *testing.T) {
		_, err := DecryptDips(":")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input text")
	})

	t.Run("invalid format - empty IV", func(t *testing.T) {
		_, err := DecryptDips(":encrypteddata")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input text")
	})

	t.Run("invalid format - empty encrypted data", func(t *testing.T) {
		_, err := DecryptDips("iv:")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid input text")
	})

	t.Run("invalid hex IV", func(t *testing.T) {
		_, err := DecryptDips("invalidhex:1234567890abcdef")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
	})

	t.Run("invalid hex encrypted data", func(t *testing.T) {
		_, err := DecryptDips("0123456789abcdef:invalidhex")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode")
	})

	t.Run("invalid cipher key", func(t *testing.T) {
		os.Setenv("DIPS_PASSWORD", "short") // invalid length

		encrypted := "0123456789abcdef0123456789abcdef:1234567890abcdef1234567890abcdef"
		_, err := DecryptDips(encrypted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create cipher block")
	})
}

func TestHMACSHA256(t *testing.T) {
	originalKey := os.Getenv("ENCRYPT_KEY")
	defer func() {
		os.Setenv("ENCRYPT_KEY", originalKey)
	}()

	t.Run("successful HMAC generation", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "test-encrypt-key")

		result, err := HMACSHA256("test message")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Len(t, result, 64) // SHA256 hex string length
	})

	t.Run("empty message", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "test-encrypt-key")

		result, err := HMACSHA256("")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Len(t, result, 64)
	})

	t.Run("missing ENCRYPT_KEY", func(t *testing.T) {
		os.Unsetenv("ENCRYPT_KEY")

		_, err := HMACSHA256("test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ENCRYPT_KEY environment variable is not set")
	})

	t.Run("consistency check", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "consistent-key")

		result1, err1 := HMACSHA256("same message")
		result2, err2 := HMACSHA256("same message")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, result1, result2)
	})

	t.Run("different messages produce different hashes", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "test-key")

		result1, err1 := HMACSHA256("message1")
		result2, err2 := HMACSHA256("message2")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, result1, result2)
	})
}

func TestHMACSHA1(t *testing.T) {
	t.Run("successful HMAC SHA1 generation", func(t *testing.T) {
		result, err := HMACSHA1("test message", "test-key")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// Check if it's valid base64
		assert.NotContains(t, result, " ")
		assert.NotContains(t, result, "\n")
	})

	t.Run("empty message", func(t *testing.T) {
		result, err := HMACSHA1("", "test-key")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("empty key", func(t *testing.T) {
		result, err := HMACSHA1("test message", "")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("consistency check", func(t *testing.T) {
		result1, err1 := HMACSHA1("same message", "same key")
		result2, err2 := HMACSHA1("same message", "same key")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, result1, result2)
	})

	t.Run("different keys produce different hashes", func(t *testing.T) {
		result1, err1 := HMACSHA1("message", "key1")
		result2, err2 := HMACSHA1("message", "key2")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, result1, result2)
	})
}

func TestEncryptAESCBC(t *testing.T) {
	originalEncryptKey := os.Getenv("ENCRYPT_KEY")
	originalIVKey := os.Getenv("IV_KEY")

	defer func() {
		os.Setenv("ENCRYPT_KEY", originalEncryptKey)
		os.Setenv("IV_KEY", originalIVKey)
	}()

	t.Run("successful encryption", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "12345678901234567890123456789012") // 32 bytes
		os.Setenv("IV_KEY", "1234567890123456")                      // 16 bytes

		result, err := EncryptAESCBC("test message")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("empty message", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "12345678901234567890123456789012")
		os.Setenv("IV_KEY", "1234567890123456")

		result, err := EncryptAESCBC("")
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("invalid ENCRYPT_KEY length", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "short") // wrong length
		os.Setenv("IV_KEY", "1234567890123456")

		_, err := EncryptAESCBC("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ENCRYPT_KEY must be 32 bytes")
	})

	t.Run("invalid IV_KEY length", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "12345678901234567890123456789012")
		os.Setenv("IV_KEY", "short") // wrong length

		_, err := EncryptAESCBC("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "IV_KEY must be 16 bytes")
	})

	t.Run("missing ENCRYPT_KEY", func(t *testing.T) {
		os.Unsetenv("ENCRYPT_KEY")
		os.Setenv("IV_KEY", "1234567890123456")

		_, err := EncryptAESCBC("test")
		assert.Error(t, err)
	})

	t.Run("missing IV_KEY", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "12345678901234567890123456789012")
		os.Unsetenv("IV_KEY")

		_, err := EncryptAESCBC("test")
		assert.Error(t, err)
	})
}

func TestDecryptAESCBC(t *testing.T) {
	originalEncryptKey := os.Getenv("ENCRYPT_KEY")
	originalIVKey := os.Getenv("IV_KEY")

	defer func() {
		os.Setenv("ENCRYPT_KEY", originalEncryptKey)
		os.Setenv("IV_KEY", originalIVKey)
	}()

	// Set valid keys for testing
	encryptKey := "12345678901234567890123456789012" // 32 bytes
	ivKey := "1234567890123456"                      // 16 bytes

	t.Run("successful encrypt and decrypt", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", encryptKey)
		os.Setenv("IV_KEY", ivKey)

		original := "test message for encryption"
		encrypted, err := EncryptAESCBC(original)
		assert.NoError(t, err)

		decrypted, err := DecryptAESCBC(encrypted)
		assert.NoError(t, err)
		assert.Equal(t, original, decrypted)
	})

	t.Run("invalid ENCRYPT_KEY length", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", "short")
		os.Setenv("IV_KEY", ivKey)

		_, err := DecryptAESCBC("dGVzdA==") // dummy base64
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ENCRYPT_KEY must be 32 bytes")
	})

	t.Run("invalid IV_KEY length", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", encryptKey)
		os.Setenv("IV_KEY", "short")

		_, err := DecryptAESCBC("dGVzdA==")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "IV_KEY must be 16 bytes")
	})

	t.Run("invalid base64", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", encryptKey)
		os.Setenv("IV_KEY", ivKey)

		_, err := DecryptAESCBC("invalid-base64!")
		assert.Error(t, err)
	})

	t.Run("invalid block size", func(t *testing.T) {
		os.Setenv("ENCRYPT_KEY", encryptKey)
		os.Setenv("IV_KEY", ivKey)

		// Create base64 of data that's not multiple of block size
		invalidData := "dGVzdA==" // "test" in base64, not multiple of 16
		_, err := DecryptAESCBC(invalidData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ciphertext is not a multiple of the block size")
	})
}

func TestPadUnpad(t *testing.T) {
	t.Run("pad function", func(t *testing.T) {
		data := []byte("hello")
		blockSize := 16

		padded := pad(data, blockSize)

		// Should be padded to next multiple of blockSize
		assert.Equal(t, 16, len(padded))
		assert.Equal(t, []byte("hello"), padded[:5])

		// Check padding bytes
		paddingValue := 16 - 5 // 11
		for i := 5; i < 16; i++ {
			assert.Equal(t, byte(paddingValue), padded[i])
		}
	})

	t.Run("pad exact block size", func(t *testing.T) {
		data := make([]byte, 16) // exactly one block
		padded := pad(data, 16)

		// Should add a full block of padding
		assert.Equal(t, 32, len(padded))

		// Last 16 bytes should all be 16
		for i := 16; i < 32; i++ {
			assert.Equal(t, byte(16), padded[i])
		}
	})

	t.Run("unpad function", func(t *testing.T) {
		// Create padded data manually
		data := []byte("hello")
		paddingValue := byte(11)
		padded := append(data, []byte{paddingValue, paddingValue, paddingValue, paddingValue, paddingValue, paddingValue, paddingValue, paddingValue, paddingValue, paddingValue, paddingValue}...)

		unpadded := unpad(padded)
		assert.Equal(t, []byte("hello"), unpadded)
	})

	t.Run("unpad empty data", func(t *testing.T) {
		result := unpad([]byte{})
		assert.Nil(t, result)
	})

	t.Run("unpad invalid padding size", func(t *testing.T) {
		// Padding size larger than data length
		data := []byte{1, 2, 255} // last byte indicates padding size of 255
		result := unpad(data)
		assert.Nil(t, result)
	})

	t.Run("unpad zero padding", func(t *testing.T) {
		// Zero padding is invalid
		data := []byte{1, 2, 3, 0}
		result := unpad(data)
		assert.Nil(t, result)
	})

	t.Run("unpad inconsistent padding", func(t *testing.T) {
		// Inconsistent padding bytes - last byte says padding is 3, but not all padding bytes are 3
		data := []byte{1, 2, 3, 2, 3} // last byte says padding is 3, but middle padding byte is 2 (inconsistent)
		result := unpad(data)
		assert.Nil(t, result)
	})

	t.Run("round trip pad/unpad", func(t *testing.T) {
		original := []byte("test message for padding")
		padded := pad(original, 16)
		unpadded := unpad(padded)

		assert.Equal(t, original, unpadded)
	})
}
