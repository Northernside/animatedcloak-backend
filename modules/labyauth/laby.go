package labyauth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strings"
	"time"
)

type LabyPayload struct {
	UUID     string   `json:"uuid"`
	Username string   `json:"user_name"`
	Roles    []string `json:"roles"`
	Issuer   string   `json:"iss"`
	Exp      int      `json:"exp"`
	Iat      int      `json:"iat"`
}

const publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt3rCKqrQYcmSEE8zyQTA
7flKIe1pr7GHY58lTF74Pw/ZZYzxmScYteXp8XBvrQfPj4U/v9Vum8IPg6GHOv1G
de3rY5ydfunEKi/w4ibVN5buPpndzcNaMoQvEJ/B5VLIzCvLc5HepFKbKFOGu8Xo
Fz8NZY0lUfGLR0rcDsHWZLHPhqYsIsUd9snkWkHaIKD7l9xTd77PpLZiBwCPnVh
h3invFY2OnCL6BfiJhhud/aDaAzFW981J9EhyACbuac2qu6Uz2bKX/7Af01gUs48
MbKUx8YirBWLD7j/CJMWorTT467It4mAvDlw43s3Py9IvxCzEFnOIftIv+7wwv1R
jVQIDAQAB
-----END PUBLIC KEY-----`

func TrimTokenType(token string) string {
	return strings.Join(strings.Split(token, " ")[1:], " ")
}

func VerifyToken(token string) (LabyPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 3 {
		return LabyPayload{}, errors.New("invalid token format")
	}

	header, err := base64UrlDecode(parts[0])
	if err != nil {
		return LabyPayload{}, err
	}

	payload, err := base64UrlDecode(parts[1])
	if err != nil {
		return LabyPayload{}, err
	}

	signature, err := base64UrlDecode(parts[2])
	if err != nil {
		return LabyPayload{}, err
	}

	var headerMap map[string]interface{}
	if err := json.Unmarshal(header, &headerMap); err != nil {
		return LabyPayload{}, err
	}
	if headerMap["alg"] != "RS256" {
		return LabyPayload{}, errors.New("unexpected signing method")
	}

	block, _ := pem.Decode([]byte(publicKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		return LabyPayload{}, errors.New("failed to parse public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return LabyPayload{}, err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return LabyPayload{}, errors.New("invalid public key type")
	}

	message := parts[0] + "." + parts[1]
	hashed := sha256.Sum256([]byte(message))

	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return LabyPayload{}, err
	}

	var payloadData LabyPayload
	if err := json.Unmarshal(payload, &payloadData); err != nil {
		return LabyPayload{}, err
	}

	if int64(payloadData.Exp) < time.Now().Unix() {
		return LabyPayload{}, errors.New("token is expired")
	}

	return payloadData, nil
}

func base64UrlDecode(input string) ([]byte, error) {
	switch len(input) % 4 {
	case 2:
		input += "=="
	case 3:
		input += "="
	}

	return base64.URLEncoding.DecodeString(input)
}
