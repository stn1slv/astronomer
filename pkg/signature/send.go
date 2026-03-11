package signature

import (
	"bytes"
	stdcontext "context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/Ullaakut/disgo"
	astronomer_context "github.com/stn1slv/astronomer/pkg/context"
	"github.com/stn1slv/astronomer/pkg/trust"
)

// SignedReport represents a report that has been signed
// by a legitimate version of Astronomer.
type SignedReport struct {
	*trust.Report

	RepositoryOwner string
	RepositoryName  string

	Signature []byte
}

// SendReport signs a report and sends it to Astrolab.
func SendReport(ctx stdcontext.Context, astronomerCtx *astronomer_context.Context, report *trust.Report) error {
	signature, err := signReport(report)
	if err != nil {
		return err
	}

	return sendReport(ctx, SignedReport{
		Report:          report,
		RepositoryOwner: astronomerCtx.RepoOwner,
		RepositoryName:  astronomerCtx.RepoName,
		Signature:       signature,
	})
}

func signReport(report *trust.Report) ([]byte, error) {
	data, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal trust report: %w", err)
	}

	hashedReport := sha512.Sum512(data)

	pemKey := os.Getenv("ASTRONOMER_PRIVATE_KEY")
	if pemKey == "" {
		pemKey = privateKeyPemData
	}

	keyBlock, _ := pem.Decode([]byte(pemKey))
	if keyBlock == nil {
		return nil, fmt.Errorf("unable to decode private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA512, hashedReport[:])
	if err != nil {
		return nil, fmt.Errorf("unable to sign trust report: %w", err)
	}

	return signature, nil
}

func sendReport(ctx stdcontext.Context, report SignedReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("unable to marshal signed report: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://astronomer.ullaakut.eu", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("unable to prepare request to astronomer server: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send signed report to astronomer server: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != 201 {
		return fmt.Errorf("astronomer server did not trust this report: %v", response.Status)
	}

	disgo.Debugln("Signed report successfully sent to astronomer server, thanks for your contribution!")

	return nil
}

var privateKeyPemData = `👀`
var publicKeyPemData = `👀`
