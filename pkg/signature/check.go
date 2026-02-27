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
		return fmt.Errorf("unable to marshal trust report: %v", err)
	}

	hashedReport := sha512.Sum512(data)

	keyBlock, _ := pem.Decode([]byte(pemData))
	if keyBlock == nil {
		return fmt.Errorf("unable to decode key")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse key: %v", err)
	}

	err = rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA512, hashedReport[:], report.Signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}

	return nil
}
