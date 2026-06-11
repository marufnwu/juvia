package auth

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/pquerna/otp/totp"
)

// TOTPManager handles TOTP 2FA operations.
type TOTPManager struct{}

// NewTOTPManager creates a new TOTPManager.
func NewTOTPManager() *TOTPManager {
	return &TOTPManager{}
}

// GenerateSecret creates a new TOTP secret and returns the secret and QR URI.
func (t *TOTPManager) GenerateSecret(issuer, accountName string) (secret string, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate totp key: %w", err)
	}
	return key.Secret(), key.URL(), nil
}

// VerifyCode validates a TOTP code against a secret.
func (t *TOTPManager) VerifyCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateRecoveryCodes generates 10 single-use recovery codes.
func (t *TOTPManager) GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("generate recovery code: %w", err)
		}
		num := binary.BigEndian.Uint32(b)
		codes[i] = fmt.Sprintf("%04d-%04d-%04d", (num>>20)&0xFFF, (num>>8)&0xFFF, num&0xFF)
	}
	return codes, nil
}

// NormalizeSecret removes spaces and uppercases the secret.
func NormalizeSecret(secret string) string {
	return strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
}

// EncodeSecret base32 encodes a raw secret.
func EncodeSecret(raw []byte) string {
	return base32.StdEncoding.EncodeToString(raw)
}
