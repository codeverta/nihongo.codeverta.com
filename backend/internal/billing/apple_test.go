package billing

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"nihongo-master-backend/internal/domain"
	"strings"
	"testing"
	"time"
)

func appleTestSigner(t *testing.T) (*StoreVerifier, func(appleTransaction) string) {
	t.Helper()
	keys := []*ecdsa.PrivateKey{}
	certs := []*x509.Certificate{}
	ders := [][]byte{}
	for i := 0; i < 3; i++ {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, key)
		template := &x509.Certificate{SerialNumber: big.NewInt(int64(i + 1)), Subject: pkix.Name{CommonName: "test certificate"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), BasicConstraintsValid: true, KeyUsage: x509.KeyUsageDigitalSignature}
		if i < 2 {
			template.IsCA = true
			template.KeyUsage |= x509.KeyUsageCertSign
		}
		if i == 1 {
			template.ExtraExtensions = []pkix.Extension{{Id: asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 2, 1}, Value: []byte{5, 0}}}
		}
		if i == 2 {
			template.ExtraExtensions = []pkix.Extension{{Id: asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}, Value: []byte{5, 0}}}
		}
		parent, signing := template, key
		if i > 0 {
			parent, signing = certs[i-1], keys[i-1]
		}
		der, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, signing)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		certs = append(certs, cert)
		ders = append(ders, der)
	}
	roots := x509.NewCertPool()
	roots.AddCert(certs[0])
	v := &StoreVerifier{roots: roots, ProductID: "lifetime", BundleID: "com.example.test", AppleEnvironment: "Production", appleKey: keys[2], AppleIssuer: "issuer", AppleKeyID: "key"}
	sign := func(transaction appleTransaction) string {
		header, _ := json.Marshal(map[string]any{"alg": "ES256", "x5c": []string{base64.StdEncoding.EncodeToString(ders[2]), base64.StdEncoding.EncodeToString(ders[1]), base64.StdEncoding.EncodeToString(ders[0])}})
		raw, _ := json.Marshal(transaction)
		var values map[string]any
		json.Unmarshal(raw, &values)
		values["signedDate"] = time.Now().UnixMilli()
		body, _ := json.Marshal(values)
		text := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(body)
		hash := sha256.Sum256([]byte(text))
		r, s, err := ecdsa.Sign(rand.Reader, keys[2], hash[:])
		if err != nil {
			t.Fatal(err)
		}
		signature := make([]byte, 64)
		r.FillBytes(signature[:32])
		s.FillBytes(signature[32:])
		return text + "." + base64.RawURLEncoding.EncodeToString(signature)
	}
	return v, sign
}

func TestAppleProofAuthenticityAndFreshRevocation(t *testing.T) {
	v, sign := appleTestSigner(t)
	transaction := appleTransaction{TransactionID: "100", OriginalID: "100", ProductID: "lifetime", BundleID: v.BundleID, Environment: "Production", Type: "Non-Consumable", Ownership: "PURCHASED", AccountToken: "buyer-account-token"}
	signed := sign(transaction)
	calls := 0
	v.appleClient = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "api.storekit.itunes.apple.com" || r.URL.Path != "/inApps/v1/transactions/100" || !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Fatal("wrong App Store Server API request")
		}
		raw, _ := json.Marshal(map[string]string{"signedTransactionInfo": sign(transaction)})
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
	})}
	proof := domain.PurchaseProof{Platform: "ios", ProductID: "lifetime", VerificationData: signed, TransactionID: "100"}
	p, err := v.Verify(context.Background(), proof)
	if err != nil || p.State != "purchased" || p.AccountToken != transaction.AccountToken {
		t.Fatalf("valid Apple proof rejected: %+v %v", p, err)
	}
	transaction.RevocationDate = time.Now().UnixMilli()
	p, err = v.Verify(context.Background(), proof)
	if err != nil || p.State != "revoked" || calls != 2 {
		t.Fatal("stale Apple receipt resurrected refund")
	}
	tampered := proof
	parts := strings.Split(signed, ".")
	parts[2] = base64.RawURLEncoding.EncodeToString(make([]byte, 64))
	tampered.VerificationData = strings.Join(parts, ".")
	if _, err = v.Verify(context.Background(), tampered); !errors.Is(err, ErrInvalid) {
		t.Fatal("forged signature accepted")
	}
	tampered = proof
	transaction.BundleID = "wrong.bundle"
	tampered.VerificationData = sign(transaction)
	if _, err = v.Verify(context.Background(), tampered); !errors.Is(err, ErrInvalid) {
		t.Fatal("wrong app accepted")
	}
	v.roots = x509.NewCertPool()
	if _, err = v.Verify(context.Background(), proof); !errors.Is(err, ErrInvalid) {
		t.Fatal("untrusted Apple signing chain accepted")
	}
}
