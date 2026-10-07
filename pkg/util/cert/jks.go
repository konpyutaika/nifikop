package cert

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	keystore "github.com/pavel-v-chernykh/keystore-go"
)

const (
	// JKSKeyAlias is the alias given to the private key entry of a generated keystore.
	// cert-manager uses the same alias for the keystores it builds, so a keystore
	// produced here is interchangeable with one produced by the cert-manager backend.
	JKSKeyAlias = "certificate"
	// JKSCAAlias is the alias given to the CA entry of a generated truststore.
	JKSCAAlias = "ca"
)

// DecodeCertificateChain returns every certificate found in a PEM bundle, in order.
func DecodeCertificateChain(raw []byte) (certs []*x509.Certificate, err error) {
	for block, rest := pem.Decode(raw); block != nil; block, rest = pem.Decode(rest) {
		if block.Type != "CERTIFICATE" {
			continue
		}
		var cert *x509.Certificate
		if cert, err = x509.ParseCertificate(block.Bytes); err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	if len(certs) == 0 {
		return nil, errors.New("failed to decode x509 certificate from PEM")
	}
	return certs, nil
}

// GenerateJKSKeyStore builds a JKS keystore holding a private key and its certificate
// chain, encrypted with the given password. Unlike GenerateJKS it does not invent a
// password, so the keystore and truststore of a secret can share one.
func GenerateJKSKeyStore(clientCert, clientKey, clientCA, password []byte) (out []byte, err error) {
	chain, err := DecodeCertificateChain(clientCert)
	if err != nil {
		return nil, err
	}

	key, err := DecodeKey(clientKey)
	if err != nil {
		return nil, err
	}

	certBundle := make([]keystore.Certificate, 0, len(chain)+1)
	for _, cert := range chain {
		certBundle = append(certBundle, keystore.Certificate{Type: "X.509", Content: cert.Raw})
	}

	// Append the CA only when the certificate itself did not already carry the chain,
	// otherwise the issuer would appear twice in the keystore.
	if len(chain) == 1 && len(clientCA) > 0 {
		var ca []*x509.Certificate
		if ca, err = DecodeCertificateChain(clientCA); err != nil {
			return nil, err
		}
		for _, cert := range ca {
			certBundle = append(certBundle, keystore.Certificate{Type: "X.509", Content: cert.Raw})
		}
	}

	jks := keystore.KeyStore{
		JKSKeyAlias: &keystore.PrivateKeyEntry{
			Entry:     keystore.Entry{CreationDate: time.Now()},
			PrivKey:   key,
			CertChain: certBundle,
		},
	}

	var outBuf bytes.Buffer
	if err = keystore.Encode(&outBuf, jks, password); err != nil {
		return nil, err
	}
	return outBuf.Bytes(), nil
}

// GenerateJKSTrustStore builds a JKS truststore holding the given CA certificate(s),
// encrypted with the given password.
func GenerateJKSTrustStore(clientCA, password []byte) (out []byte, err error) {
	chain, err := DecodeCertificateChain(clientCA)
	if err != nil {
		return nil, err
	}

	jks := keystore.KeyStore{}
	for i, cert := range chain {
		alias := JKSCAAlias
		if i > 0 {
			// Keep additional issuers addressable rather than overwriting the first entry.
			alias = fmt.Sprintf("%s-%d", JKSCAAlias, i)
		}
		jks[alias] = &keystore.TrustedCertificateEntry{
			Entry:       keystore.Entry{CreationDate: time.Now()},
			Certificate: keystore.Certificate{Type: "X.509", Content: cert.Raw},
		}
	}

	var outBuf bytes.Buffer
	if err = keystore.Encode(&outBuf, jks, password); err != nil {
		return nil, err
	}
	return outBuf.Bytes(), nil
}
