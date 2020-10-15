package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

var (
	// set with Makefile during compilation
	detaSignVersion string
)

// CalcSignatureInput input to CalcSignature function
type CalcSignatureInput struct {
	AccessToken string
	HTTPMethod  string
	URL         string
	Timestamp   string
	ContentType string
	RawBody     []byte
}

// CalcSignature calculates the signature for signing the requests
func (m *Manager) CalcSignature(i *CalcSignatureInput) (string, error) {
	// only v0 for now
	if detaSignVersion != "v0" {
		return "", nil
	}

	tokenParts := strings.Split(i.AccessToken, "_")
	if len(tokenParts) != 2 {
		return "", ErrInvalidAccessToken
	}
	accessKeyID := tokenParts[0]
	accessKeySecret := tokenParts[1]

	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n",
		i.HTTPMethod,
		i.URL,
		i.Timestamp,
		i.ContentType,
		i.RawBody,
	)

	mac := hmac.New(sha256.New, []byte(accessKeySecret))
	_, err := mac.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}
	signature := mac.Sum(nil)
	hexSign := hex.EncodeToString(signature)

	return fmt.Sprintf("%s=%s:%s", detaSignVersion, accessKeyID, hexSign), nil
}
