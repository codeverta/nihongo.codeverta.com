package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"google.golang.org/api/idtoken"
	"net/http"
	"nihongo-master-backend/internal/billing"
	"nihongo-master-backend/internal/domain"
	"strings"
)

func (s *Server) ConfigureBilling(service *billing.Service, audience, email string) {
	s.billing = service
	s.billingPushAudience = audience
	s.billingPushEmail = email
}
func (s *Server) billingStatus(w http.ResponseWriter, r *http.Request) {
	if !s.personalBillingAccount(w, r) {
		return
	}
	token, err := s.store.BillingAccountToken(userID(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	platforms := []string{}
	product := "nihongo_master_lifetime"
	if s.billing != nil {
		platforms = s.billing.Verifier.Platforms()
		product = s.billing.ProductID
		if err = s.billing.RefreshUser(r.Context(), userID(r)); err != nil {
			writeBillingError(w, err)
			return
		}
	}
	writeData(w, 200, map[string]any{"productId": product, "accountToken": token, "priceIdr": 99000, "billingType": "lifetime", "platforms": platforms, "entitlement": s.store.Profile(userID(r)).Entitlement})
}
func (s *Server) verifyPurchase(w http.ResponseWriter, r *http.Request) {
	if !s.personalBillingAccount(w, r) {
		return
	}
	if s.billing == nil {
		writeBillingError(w, billing.ErrUnavailable)
		return
	}
	var input domain.PurchaseProof
	if decodeJSON(r, &input) != nil {
		writeError(w, 400, "invalid_json", "Data pembelian tidak valid.")
		return
	}
	p, err := s.billing.Verify(r.Context(), userID(r), input)
	if err != nil {
		writeBillingError(w, err)
		return
	}
	status := 200
	if p.State == "pending" {
		status = 202
	}
	writeData(w, status, map[string]any{"state": p.State, "entitlement": s.store.Profile(userID(r)).Entitlement})
}
func (s *Server) personalBillingAccount(w http.ResponseWriter, r *http.Request) bool {
	if s.store.IsDemoUser(userID(r)) || s.store.IsPlayReviewUser(userID(r)) {
		writeError(w, 403, "personal_account_required", "Masuk dengan akun Google pribadi untuk membeli atau restore. Akun demo/reviewer digunakan bersama.")
		return false
	}
	return true
}
func writeBillingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, billing.ErrUnavailable):
		writeError(w, 503, "billing_unavailable", "Verifikasi store belum tersedia. Coba lagi atau restore pembelian.")
	case errors.Is(err, billing.ErrInvalid):
		writeError(w, 422, "invalid_purchase", "Bukti pembelian tidak valid untuk aplikasi ini.")
	default:
		writeStoreError(w, err)
	}
}
func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Confirmation string `json:"confirmation"`
	}
	if decodeJSON(r, &input) != nil || input.Confirmation != "DELETE" {
		writeError(w, 422, "confirmation_required", "Konfirmasi penghapusan akun diperlukan.")
		return
	}
	if err := s.store.DeleteAccount(userID(r)); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) appleBillingNotification(w http.ResponseWriter, r *http.Request) {
	if s.billing == nil {
		writeBillingError(w, billing.ErrUnavailable)
		return
	}
	var input struct {
		Signed string `json:"signedPayload"`
	}
	if decodeJSON(r, &input) != nil || len(input.Signed) > 100000 {
		writeError(w, 400, "invalid_notification", "Invalid notification.")
		return
	}
	if err := s.billing.AppleNotification(r.Context(), input.Signed); err != nil {
		writeBillingError(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) googleBillingNotification(w http.ResponseWriter, r *http.Request) {
	if s.billing == nil || s.billingPushAudience == "" || s.billingPushEmail == "" {
		writeBillingError(w, billing.ErrUnavailable)
		return
	}
	claims, err := idtoken.Validate(r.Context(), bearerToken(r), s.billingPushAudience)
	if err != nil || claims.Claims["email"] != s.billingPushEmail || claims.Claims["email_verified"] != true {
		writeError(w, 401, "invalid_push_identity", "Invalid Google Pub/Sub identity.")
		return
	}
	// Pub/Sub also includes subscription/message metadata; accept the official envelope.
	var envelope struct {
		Message struct {
			Data string `json:"data"`
		} `json:"message"`
		Subscription string `json:"subscription"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if decoder.Decode(&envelope) != nil {
		writeError(w, 400, "invalid_notification", "Invalid notification.")
		return
	}
	raw, err := base64.StdEncoding.DecodeString(envelope.Message.Data)
	if err != nil {
		writeError(w, 400, "invalid_notification", "Invalid payload.")
		return
	}
	var data struct {
		Package string `json:"packageName"`
		OneTime *struct {
			Token string `json:"purchaseToken"`
			SKU   string `json:"sku"`
		} `json:"oneTimeProductNotification"`
		Voided *struct {
			Token string `json:"purchaseToken"`
		} `json:"voidedPurchaseNotification"`
		Test json.RawMessage `json:"testNotification"`
	}
	if json.Unmarshal(raw, &data) != nil {
		writeError(w, 400, "invalid_notification", "Invalid payload.")
		return
	}
	v, ok := s.billing.Verifier.(*billing.StoreVerifier)
	if !ok || data.Package != v.PackageID {
		writeError(w, 422, "invalid_notification", "Wrong package.")
		return
	}
	token := ""
	if data.OneTime != nil && data.OneTime.SKU == s.billing.ProductID {
		token = data.OneTime.Token
	}
	if data.Voided != nil {
		token = data.Voided.Token
	}
	if token != "" {
		if err = s.billing.RefreshPurchase(r.Context(), "android", token); err != nil {
			writeBillingError(w, err)
			return
		}
	}
	w.WriteHeader(204)
}
func (s *Server) authorizePremium(w http.ResponseWriter, r *http.Request, id string) bool {
	level := r.PathValue("level")
	if level == "" {
		level = r.URL.Query().Get("level")
	}
	if level == "" {
		level = "N5"
	}
	if !s.store.PremiumRequest(r.URL.Path, strings.ToUpper(strings.TrimSpace(level)), r.PathValue("lessonID"), r.PathValue("attemptID")) {
		return true
	}
	if s.billing != nil {
		if err := s.billing.RefreshUser(r.Context(), id); err != nil {
			writeBillingError(w, err)
			return false
		}
	}
	if s.store.Profile(id).Entitlement.Tier != domain.TierPremium {
		writeError(w, 402, "premium_required", "Materi N3–N1 dan tes selain N5 memerlukan Lifetime Access.")
		return false
	}
	return true
}
