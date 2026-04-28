package abacatepay_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"facilitador-de-doacoes/pkg/abacatepay"
)

func computeHMAC(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestValidateSignature_Valid(t *testing.T) {
	secret := "webhook-secret"
	body := []byte(`{"eventType":"billing.paid"}`)
	sig := computeHMAC(secret, body)

	assert.True(t, abacatepay.ValidateSignature(secret, sig, body))
}

func TestValidateSignature_InvalidSignature(t *testing.T) {
	body := []byte(`{"eventType":"billing.paid"}`)

	assert.False(t, abacatepay.ValidateSignature("secret", "bad-signature", body))
}

func TestValidateSignature_WrongSecret(t *testing.T) {
	body := []byte(`{"eventType":"billing.paid"}`)
	sig := computeHMAC("correct-secret", body)

	assert.False(t, abacatepay.ValidateSignature("wrong-secret", sig, body))
}

func TestValidateSignature_TamperedBody(t *testing.T) {
	secret := "my-secret"
	originalBody := []byte(`{"amount":1000}`)
	sig := computeHMAC(secret, originalBody)

	tamperedBody := []byte(`{"amount":9999}`)
	assert.False(t, abacatepay.ValidateSignature(secret, sig, tamperedBody))
}

func TestValidateSignature_EmptyBody(t *testing.T) {
	secret := "my-secret"
	body := []byte("")
	sig := computeHMAC(secret, body)

	assert.True(t, abacatepay.ValidateSignature(secret, sig, body))
}
