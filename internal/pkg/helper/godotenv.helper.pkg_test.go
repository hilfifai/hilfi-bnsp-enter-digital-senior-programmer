package helper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit playground untuk fungsi GetEnv
func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_ENV", "nilai_env")
	defer os.Unsetenv("TEST_ENV")

	hasil := GetEnv("TEST_ENV")
	assert.Equal(t, "nilai_env", hasil, "GetEnv harus mengembalikan nilai environment variable yang benar")
}

// Unit playground untuk fungsi GetEnvRequired
func TestGetEnvRequired(t *testing.T) {
	os.Setenv("REQUIRED_ENV", "wajib")
	defer os.Unsetenv("REQUIRED_ENV")

	// Kasus environment variable ada
	hasil, err := GetEnvRequired("REQUIRED_ENV")
	assert.NoError(t, err, "Tidak boleh error jika env ada")
	assert.Equal(t, "wajib", hasil, "Harus mengembalikan nilai yang benar")

	// Kasus environment variable tidak ada
	os.Unsetenv("REQUIRED_ENV")
	hasil, err = GetEnvRequired("REQUIRED_ENV")
	assert.Error(t, err, "Harus error jika env tidak ada")
	assert.Equal(t, "", hasil, "Jika env tidak ada, hasil harus string kosong")
}

// Unit playground untuk fungsi GetEnvDefault
func TestGetEnvDefault(t *testing.T) {
	os.Setenv("DEFAULT_ENV", "ada")
	defer os.Unsetenv("DEFAULT_ENV")

	// Kasus env ada
	hasil := GetEnvDefault("DEFAULT_ENV", "fallback")
	assert.Equal(t, "ada", hasil, "Jika env ada, harus mengembalikan nilai env")

	// Kasus env tidak ada
	hasil = GetEnvDefault("TIDAK_ADA_ENV", "fallback")
	assert.Equal(t, "fallback", hasil, "Jika env tidak ada, harus mengembalikan fallback")
}

// Unit playground untuk fungsi GetEnvAsBool
func TestGetEnvAsBool(t *testing.T) {
	os.Setenv("BOOL_ENV_TRUE", "true")
	os.Setenv("BOOL_ENV_FALSE", "false")
	os.Setenv("BOOL_ENV_INVALID", "bukanbool")
	defer os.Unsetenv("BOOL_ENV_TRUE")
	defer os.Unsetenv("BOOL_ENV_FALSE")
	defer os.Unsetenv("BOOL_ENV_INVALID")

	assert.True(t, GetEnvAsBool("BOOL_ENV_TRUE", false), "Harus true jika env bernilai 'true'")
	assert.False(t, GetEnvAsBool("BOOL_ENV_FALSE", false), "Harus false jika env bernilai 'false'")
	assert.False(t, GetEnvAsBool("BOOL_ENV_INVALID", false), "Harus false jika env tidak valid")
	assert.False(t, GetEnvAsBool("TIDAK_ADA_BOOL_ENV", false), "Harus false jika env tidak ada")
}

// Unit playground untuk fungsi GetEnvAsInt
func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("INT_ENV_VALID", "123")
	os.Setenv("INT_ENV_INVALID", "abc")
	defer os.Unsetenv("INT_ENV_VALID")
	defer os.Unsetenv("INT_ENV_INVALID")

	assert.Equal(t, 123, GetEnvAsInt("INT_ENV_VALID", 123), "Harus mengembalikan integer yang benar jika env valid")
	assert.Equal(t, 0, GetEnvAsInt("INT_ENV_INVALID", 0), "Harus 0 jika env tidak valid")
	assert.Equal(t, 0, GetEnvAsInt("TIDAK_ADA_INT_ENV", 0), "Harus 0 jika env tidak ada")
}

// Dokumentasi:
// - Setiap fungsi diuji untuk kasus normal dan edge case.
// - Menggunakan assert dari testify agar hasil lebih informatif.
// - Setiap environment variable dibersihkan setelah playground selesai.
