package billing

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/oauth2/google"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"nihongo-master-backend/internal/domain"
	"os"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid store purchase")
var ErrUnavailable = errors.New("store verification unavailable")

//go:embed AppleRootCA-G3.cer
var appleRoot []byte

type Verifier interface {
	Verify(context.Context, domain.PurchaseProof) (domain.Purchase, error)
	Acknowledge(context.Context, domain.Purchase) error
	Platforms() []string
}

type StoreVerifier struct {
	ProductID, PackageID, BundleID, AppleIssuer, AppleKeyID, AppleEnvironment string
	appleKey                                                                  *ecdsa.PrivateKey
	googleClient, appleClient                                                 *http.Client
	roots                                                                     *x509.CertPool
}

func NewStoreVerifier(ctx context.Context) (*StoreVerifier, error) {
	v := &StoreVerifier{ProductID: os.Getenv("LIFETIME_PRODUCT_ID"), PackageID: os.Getenv("ANDROID_PACKAGE_ID"), BundleID: os.Getenv("APPLE_BUNDLE_ID"), AppleIssuer: os.Getenv("APPLE_ISSUER_ID"), AppleKeyID: os.Getenv("APPLE_KEY_ID"), AppleEnvironment: os.Getenv("APPLE_ENVIRONMENT"), appleClient: &http.Client{Timeout: 12 * time.Second}, roots: x509.NewCertPool()}
	if v.ProductID == "" {
		v.ProductID = "nihongo_master_lifetime"
	}
	if v.PackageID == "" {
		v.PackageID = "com.codeverta.nihongomaster"
	}
	if v.AppleEnvironment == "" {
		v.AppleEnvironment = "Production"
	}
	if v.AppleEnvironment != "Production" && v.AppleEnvironment != "Sandbox" {
		return nil, fmt.Errorf("APPLE_ENVIRONMENT must be Production or Sandbox")
	}
	cert, err := x509.ParseCertificate(appleRoot)
	if err != nil {
		return nil, err
	}
	v.roots.AddCert(cert)
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" || os.Getenv("GOOGLE_PLAY_USE_ADC") == "true" {
		v.googleClient, err = google.DefaultClient(ctx, "https://www.googleapis.com/auth/androidpublisher")
		if err != nil {
			return nil, fmt.Errorf("Google Play credentials: %w", err)
		}
		v.googleClient.Timeout = 12 * time.Second
	}
	if path := os.Getenv("APPLE_PRIVATE_KEY_PATH"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(data)
		if block == nil {
			return nil, fmt.Errorf("invalid Apple key PEM")
		}
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		v.appleKey, ok = key.(*ecdsa.PrivateKey)
		if !ok || v.appleKey.Curve.Params().BitSize != 256 {
			return nil, fmt.Errorf("Apple key must be ES256")
		}
		if v.BundleID == "" || v.AppleIssuer == "" || v.AppleKeyID == "" {
			return nil, fmt.Errorf("Apple bundle, issuer and key IDs are required")
		}
	}
	return v, nil
}
func (v *StoreVerifier) Platforms() []string {
	p := []string{}
	if v.googleClient != nil {
		p = append(p, "android")
	}
	if v.appleKey != nil {
		p = append(p, "ios")
	}
	return p
}

func (v *StoreVerifier) Verify(ctx context.Context, proof domain.PurchaseProof) (domain.Purchase, error) {
	if proof.ProductID != v.ProductID || len(proof.VerificationData) > 100000 || proof.VerificationData == "" {
		return domain.Purchase{}, ErrInvalid
	}
	switch proof.Platform {
	case "android":
		return v.verifyGoogle(ctx, proof)
	case "ios":
		return v.verifyApple(ctx, proof)
	}
	return domain.Purchase{}, ErrInvalid
}

func (v *StoreVerifier) verifyGoogle(ctx context.Context, proof domain.PurchaseProof) (domain.Purchase, error) {
	if v.googleClient == nil {
		return domain.Purchase{}, ErrUnavailable
	}
	endpoint := "https://androidpublisher.googleapis.com/androidpublisher/v3/applications/" + url.PathEscape(v.PackageID) + "/purchases/productsv2/tokens/" + url.PathEscape(proof.VerificationData)
	var data struct {
		Acknowledgement string `json:"acknowledgementState"`
		ProductLineItem []struct {
			ProductID string `json:"productId"`
			Offer     struct {
				Quantity           int             `json:"quantity"`
				RefundableQuantity int             `json:"refundableQuantity"`
				ConsumptionState   string          `json:"consumptionState"`
				Rent               json.RawMessage `json:"rentOfferDetails"`
			} `json:"productOfferDetails"`
		} `json:"productLineItem"`
		State struct {
			PurchaseState string `json:"purchaseState"`
		} `json:"purchaseStateContext"`
		Account string `json:"obfuscatedExternalAccountId"`
	}
	if err := requestJSON(ctx, v.googleClient, http.MethodGet, endpoint, "", nil, &data); err != nil {
		return domain.Purchase{}, err
	}
	if len(data.ProductLineItem) != 1 || data.ProductLineItem[0].ProductID != v.ProductID || data.Account == "" {
		return domain.Purchase{}, ErrInvalid
	}
	item := data.ProductLineItem[0]
	if len(item.Offer.Rent) > 0 && string(item.Offer.Rent) != "null" {
		return domain.Purchase{}, ErrInvalid
	}
	state := "revoked"
	switch data.State.PurchaseState {
	case "PURCHASED":
		if item.Offer.RefundableQuantity > 0 && item.Offer.ConsumptionState == "CONSUMPTION_STATE_YET_TO_BE_CONSUMED" {
			state = "purchased"
		}
	case "PENDING":
		state = "pending"
	case "CANCELLED":
		state = "revoked"
	default:
		return domain.Purchase{}, ErrInvalid
	}
	// Google purchase tokens are unguessable identities and remain unconsumed.
	return domain.Purchase{Acknowledged: data.Acknowledgement == "ACKNOWLEDGEMENT_STATE_ACKNOWLEDGED", Platform: "android", ID: proof.VerificationData, ProductID: v.ProductID, AccountToken: data.Account, Proof: proof.VerificationData, State: state, CheckedAt: time.Now().UTC()}, nil
}

func (v *StoreVerifier) Acknowledge(ctx context.Context, p domain.Purchase) error {
	if p.Platform != "android" || p.State != "purchased" || p.Acknowledged {
		return nil
	}
	endpoint := "https://androidpublisher.googleapis.com/androidpublisher/v3/applications/" + url.PathEscape(v.PackageID) + "/purchases/products/" + url.PathEscape(v.ProductID) + "/tokens/" + url.PathEscape(p.ID) + ":acknowledge"
	return requestJSON(ctx, v.googleClient, http.MethodPost, endpoint, "", strings.NewReader(`{}`), nil)
}

type appleTransaction struct {
	TransactionID  string `json:"transactionId"`
	OriginalID     string `json:"originalTransactionId"`
	ProductID      string `json:"productId"`
	BundleID       string `json:"bundleId"`
	AccountToken   string `json:"appAccountToken"`
	Type           string `json:"type"`
	Environment    string `json:"environment"`
	RevocationDate int64  `json:"revocationDate"`
	Ownership      string `json:"inAppOwnershipType"`
}

func (v *StoreVerifier) verifyApple(ctx context.Context, proof domain.PurchaseProof) (domain.Purchase, error) {
	if v.appleKey == nil {
		return domain.Purchase{}, ErrUnavailable
	}
	var original appleTransaction
	if err := v.decodeAppleJWS(proof.VerificationData, &original); err != nil {
		return domain.Purchase{}, err
	}
	if original.TransactionID == "" || original.AccountToken == "" || original.Environment != v.AppleEnvironment || original.BundleID != v.BundleID || original.ProductID != v.ProductID || original.Type != "Non-Consumable" || original.Ownership != "PURCHASED" {
		return domain.Purchase{}, ErrInvalid
	}
	if proof.TransactionID != "" && proof.TransactionID != original.TransactionID {
		return domain.Purchase{}, ErrInvalid
	}
	token, err := v.appleJWT()
	if err != nil {
		return domain.Purchase{}, err
	}
	host := "https://api.storekit.itunes.apple.com"
	if v.AppleEnvironment == "Sandbox" {
		host = "https://api.storekit-sandbox.itunes.apple.com"
	}
	var response struct {
		Signed string `json:"signedTransactionInfo"`
	}
	if err = requestJSON(ctx, v.appleClient, http.MethodGet, host+"/inApps/v1/transactions/"+url.PathEscape(original.TransactionID), token, nil, &response); err != nil {
		return domain.Purchase{}, err
	}
	var current appleTransaction
	if err = v.decodeAppleJWS(response.Signed, &current); err != nil {
		return domain.Purchase{}, err
	}
	if current.TransactionID != original.TransactionID || current.OriginalID == "" || current.AccountToken != original.AccountToken || current.BundleID != v.BundleID || current.ProductID != v.ProductID || current.Type != "Non-Consumable" || current.Environment != v.AppleEnvironment || current.Ownership != "PURCHASED" {
		return domain.Purchase{}, ErrInvalid
	}
	state := "purchased"
	if current.RevocationDate > 0 {
		state = "revoked"
	}
	return domain.Purchase{Platform: "ios", ID: current.OriginalID, ProductID: v.ProductID, AccountToken: current.AccountToken, Proof: response.Signed, State: state, CheckedAt: time.Now().UTC()}, nil
}

// Verify the Apple certificate chain, signing extensions and ES256 signature.
// Current ownership/revocation is then read from the authenticated Server API,
// so a previously signed receipt cannot resurrect a refunded purchase.
func (v *StoreVerifier) decodeAppleJWS(signed string, result any) error {
	parts := strings.Split(signed, ".")
	if len(parts) != 3 {
		return ErrInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrInvalid
	}
	var header struct {
		Alg   string   `json:"alg"`
		Chain []string `json:"x5c"`
	}
	if json.Unmarshal(raw, &header) != nil || header.Alg != "ES256" || len(header.Chain) != 3 {
		return ErrInvalid
	}
	certs := []*x509.Certificate{}
	for _, encoded := range header.Chain {
		der, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return ErrInvalid
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return ErrInvalid
		}
		certs = append(certs, cert)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ErrInvalid
	}
	var date struct {
		SignedDate int64 `json:"signedDate"`
	}
	if json.Unmarshal(payload, &date) != nil || date.SignedDate <= 0 || date.SignedDate > time.Now().Add(time.Minute).UnixMilli() {
		return ErrInvalid
	}
	intermediates := x509.NewCertPool()
	intermediates.AddCert(certs[1])
	if _, err = certs[0].Verify(x509.VerifyOptions{Roots: v.roots, Intermediates: intermediates, CurrentTime: time.UnixMilli(date.SignedDate), KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny}}); err != nil {
		return ErrInvalid
	}
	if !hasExtension(certs[0], asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}) || !hasExtension(certs[1], asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 2, 1}) {
		return ErrInvalid
	}
	key, ok := certs[0].PublicKey.(*ecdsa.PublicKey)
	if !ok || key.Curve.Params().BitSize != 256 {
		return ErrInvalid
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(signature) != 64 {
		return ErrInvalid
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if !ecdsa.Verify(key, hash[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])) {
		return ErrInvalid
	}
	if json.Unmarshal(payload, result) != nil {
		return ErrInvalid
	}
	return nil
}
func hasExtension(cert *x509.Certificate, oid asn1.ObjectIdentifier) bool {
	for _, e := range cert.Extensions {
		if e.Id.Equal(oid) {
			return true
		}
	}
	return false
}

func (v *StoreVerifier) appleJWT() (string, error) {
	now := time.Now().Unix()
	header, _ := json.Marshal(map[string]string{"alg": "ES256", "kid": v.AppleKeyID, "typ": "JWT"})
	body, _ := json.Marshal(map[string]any{"iss": v.AppleIssuer, "iat": now, "exp": now + 300, "aud": "appstoreconnect-v1", "bid": v.BundleID})
	value := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(body)
	hash := sha256.Sum256([]byte(value))
	r, s, err := ecdsa.Sign(rand.Reader, v.appleKey, hash[:])
	if err != nil {
		return "", err
	}
	signature := make([]byte, 64)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	return value + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func requestJSON(ctx context.Context, client *http.Client, method, endpoint, bearer string, body io.Reader, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return ErrUnavailable
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode == 404 || response.StatusCode == 410 {
		return ErrInvalid
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ErrUnavailable
	}
	if result != nil && json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(result) != nil {
		return ErrUnavailable
	}
	return nil
}
