package vault

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	lineWrapWidth = 80
	saltSize      = 16
	nonceSize     = 12
	paramsSize    = 12
	keySize       = 32 // AES-256
)

var (
	magicNumber      = []byte{'O', 'R', 'K', 'V'}
	ErrInvalidFormat = errors.New("invalid vault file format")
	ErrInvalidAuth   = errors.New("authentication failed: wrong password or corrupted data")
)

// VaultParams represents Argon2 parameters
type VaultParams struct {
	TimeCost    uint32
	MemoryCost  uint32
	Parallelism uint32
}

// deriveKey derives an encryption key from a password using Argon2id
func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, keySize)
}

// encrypt encrypts data using AES-256-GCM with the given password
func encrypt(data []byte, password string) (string, error) {
	// Set Argon2 parameters
	params := VaultParams{
		TimeCost:    3,
		MemoryCost:  64 * 1024,
		Parallelism: 4,
	}

	// Compress data
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return "", fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip writer: %w", err)
	}
	dataToEncrypt := buf.Bytes()

	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Derive key
	key := deriveKey(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nil, nonce, dataToEncrypt, nil)

	// Encode params to binary
	paramsBuf := make([]byte, paramsSize)
	binary.LittleEndian.PutUint32(paramsBuf[0:4], params.TimeCost)
	binary.LittleEndian.PutUint32(paramsBuf[4:8], params.MemoryCost)
	binary.LittleEndian.PutUint32(paramsBuf[8:12], params.Parallelism)

	// Concatenate: magic + salt + nonce + params + ciphertext
	var blob bytes.Buffer
	blob.Write(magicNumber)
	blob.Write(salt)
	blob.Write(nonce)
	blob.Write(paramsBuf)
	blob.Write(ciphertext)

	// Base64 encode entire blob
	encoded := base64.StdEncoding.EncodeToString(blob.Bytes())

	// Wrap to 80 characters
	wrapped := wrapText(encoded, lineWrapWidth)

	return wrapped, nil
}

// decrypt decrypts vault data using the given password
func decrypt(data string, password string) (map[string]string, error) {
	// Unwrap (remove line breaks)
	unwrapped := unwrapText(data)

	// Base64 decode
	blob, err := base64.StdEncoding.DecodeString(unwrapped)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vault data: %w", err)
	}

	// Verify minimum length (magic + salt + nonce + params)
	minLength := 4 + saltSize + nonceSize + paramsSize + 16 // 16 for GCM tag minimum
	if len(blob) < minLength {
		return nil, ErrInvalidFormat
	}

	offset := 0

	// Read magic number
	if len(blob) < offset+4 {
		return nil, ErrInvalidFormat
	}
	if !bytes.Equal(blob[offset:offset+4], magicNumber) {
		return nil, ErrInvalidFormat
	}
	offset += 4

	// Read salt
	if len(blob) < offset+saltSize {
		return nil, ErrInvalidFormat
	}
	salt := blob[offset : offset+saltSize]
	offset += saltSize

	// Read nonce
	if len(blob) < offset+nonceSize {
		return nil, ErrInvalidFormat
	}
	nonce := blob[offset : offset+nonceSize]
	offset += nonceSize

	// Read params
	if len(blob) < offset+paramsSize {
		return nil, ErrInvalidFormat
	}
	paramsBuf := blob[offset : offset+paramsSize]
	offset += paramsSize

	// Read ciphertext (rest of blob)
	ciphertext := blob[offset:]

	// Decode params
	var params VaultParams
	params.TimeCost = binary.LittleEndian.Uint32(paramsBuf[0:4])
	params.MemoryCost = binary.LittleEndian.Uint32(paramsBuf[4:8])
	params.Parallelism = binary.LittleEndian.Uint32(paramsBuf[8:12])

	// Derive key
	key := deriveKey(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrInvalidAuth
	}

	// Decompress data
	gz, err := gzip.NewReader(bytes.NewReader(plaintext))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	decompressed, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip reader: %w", err)
	}

	// Parse JSON
	var vaultData map[string]string
	if err := json.Unmarshal(decompressed, &vaultData); err != nil {
		return nil, fmt.Errorf("failed to parse vault data: %w", err)
	}

	return vaultData, nil
}

// wrapText wraps text to specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	for i := 0; i < len(text); i += width {
		end := i + width
		if end > len(text) {
			end = len(text)
		}
		result.WriteString(text[i:end])
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// unwrapText removes newlines from wrapped text
func unwrapText(text string) string {
	text = strings.ReplaceAll(text, "\r", "")
	return strings.ReplaceAll(text, "\n", "")
}
