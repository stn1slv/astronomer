// Package signature provides functions to sign and verify trust reports.
package signature

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
)

// Check verifies whether a report has been signed
// with a legitimate version of Astronomer.
func Check(report *SignedReport) error {
	data, err := json.Marshal(report.Report)
	if err != nil {
		return fmt.Errorf("unable to marshal trust report: %w", err)
	}

	hashedReport := sha512.Sum512(data)

	keyBlock, _ := pem.Decode([]byte(publicKeyPemData))
	if keyBlock == nil {
		return fmt.Errorf("unable to decode public key")
	}

	key, err := x509.ParsePKIXPublicKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse public key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	err = rsa.VerifyPKCS1v15(rsaKey, crypto.SHA512, hashedReport[:], report.Signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}
