package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

func TestPricingAdminPersistsAndFeedsAppEndpoints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "pricing.json")
	memory := store.NewMemoryStore()
	server := api.NewServer(memory, fakeGoogleVerifier{})
	if err := server.ConfigurePricing(path, "pricing-secret"); err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()

	if got := getLifetimePrice(t, handler, "/v1/pricing"); got != 99000 {
		t.Fatalf("default public price = %d, want 99000", got)
	}
	request := httptest.NewRequest(http.MethodPut, "/v1/admin/pricing", strings.NewReader(`{"lifetimePriceIdr":149000}`))
	request.Header.Set("Authorization", "Bearer pricing-secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("admin update status = %d: %s", response.Code, response.Body.String())
	}

	for _, endpoint := range []string{"/v1/pricing", "/v1/plans", "/v1/home"} {
		if got := getLifetimePrice(t, handler, endpoint); got != 149000 {
			t.Errorf("%s lifetime price = %d, want 149000", endpoint, got)
		}
	}

	buyer, err := memory.AuthenticateGoogle(domain.GoogleIdentity{
		Subject: "pricing-buyer", Email: "pricing@example.com", DisplayName: "Pricing Buyer",
	})
	if err != nil {
		t.Fatal(err)
	}
	token, err := memory.CreateSession(buyer.ID)
	if err != nil {
		t.Fatal(err)
	}
	billingRequest := httptest.NewRequest(http.MethodGet, "/v1/billing", nil)
	billingRequest.Header.Set("Authorization", "Bearer "+token)
	billingResponse := httptest.NewRecorder()
	handler.ServeHTTP(billingResponse, billingRequest)
	if billingResponse.Code != http.StatusOK {
		t.Fatalf("billing status = %d: %s", billingResponse.Code, billingResponse.Body.String())
	}
	var billingPayload struct {
		Data struct {
			PriceIDR int `json:"priceIdr"`
		} `json:"data"`
	}
	if err := json.Unmarshal(billingResponse.Body.Bytes(), &billingPayload); err != nil {
		t.Fatal(err)
	}
	if billingPayload.Data.PriceIDR != 149000 {
		t.Fatalf("billing price = %d, want 149000", billingPayload.Data.PriceIDR)
	}

	restarted := api.NewServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	if err := restarted.ConfigurePricing(path, "pricing-secret"); err != nil {
		t.Fatal(err)
	}
	if got := getLifetimePrice(t, restarted.Handler(), "/v1/pricing"); got != 149000 {
		t.Fatalf("price after backend restart = %d, want 149000", got)
	}
}

func TestPricingAdminRejectsUnauthorizedAndInvalidPrices(t *testing.T) {
	server := api.NewServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	if err := server.ConfigurePricing(filepath.Join(t.TempDir(), "pricing.json"), "pricing-secret"); err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()

	for _, test := range []struct {
		name, token, body string
		wantStatus        int
	}{
		{name: "missing admin key", body: `{"lifetimePriceIdr":149000}`, wantStatus: http.StatusUnauthorized},
		{name: "too low", token: "pricing-secret", body: `{"lifetimePriceIdr":999}`, wantStatus: http.StatusUnprocessableEntity},
		{name: "too high", token: "pricing-secret", body: `{"lifetimePriceIdr":100000001}`, wantStatus: http.StatusUnprocessableEntity},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/v1/admin/pricing", strings.NewReader(test.body))
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d: %s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
	if got := getLifetimePrice(t, handler, "/v1/pricing"); got != 99000 {
		t.Fatalf("invalid update changed public price to %d", got)
	}
}

func TestLifetimeOfferSettingsPersistAndKeepOnPriceOnlySave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pricing.json")
	server := api.NewServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	if err := server.ConfigurePricing(path, "pricing-secret"); err != nil {
		t.Fatal(err)
	}
	handler := server.Handler()
	offer := func(h http.Handler) (bool, int) {
		request := httptest.NewRequest(http.MethodGet, "/v1/pricing", nil)
		response := httptest.NewRecorder()
		h.ServeHTTP(response, request)
		var payload struct {
			Data struct {
				OfferAfterLogin   bool `json:"offerAfterLogin"`
				OfferIntervalDays int  `json:"offerIntervalDays"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		return payload.Data.OfferAfterLogin, payload.Data.OfferIntervalDays
	}
	if on, days := offer(handler); !on || days != 10 {
		t.Fatalf("default offer = %v/%d, want on/10", on, days)
	}
	put := func(body string) int {
		request := httptest.NewRequest(http.MethodPut, "/v1/admin/pricing", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer pricing-secret")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Code
	}
	if code := put(`{"lifetimePriceIdr":99000,"offerAfterLogin":false,"offerIntervalDays":3}`); code != http.StatusOK {
		t.Fatalf("save offer = %d", code)
	}
	// The standalone pricing page sends only the price.
	if code := put(`{"lifetimePriceIdr":129000}`); code != http.StatusOK {
		t.Fatalf("price-only save = %d", code)
	}
	if code := put(`{"lifetimePriceIdr":99000,"offerIntervalDays":91}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("interval 91 accepted: %d", code)
	}
	restarted := api.NewServer(store.NewMemoryStore(), fakeGoogleVerifier{})
	if err := restarted.ConfigurePricing(path, "pricing-secret"); err != nil {
		t.Fatal(err)
	}
	if on, days := offer(restarted.Handler()); on || days != 3 {
		t.Fatalf("offer after price-only save and restart = %v/%d, want off/3", on, days)
	}
}

func TestProfileSaysWhetherTheAccountCanPurchase(t *testing.T) {
	memory := store.NewMemoryStore()
	buyer, err := memory.AuthenticateGoogle(domain.GoogleIdentity{Subject: "p", Email: "p@example.com", DisplayName: "P"})
	if err != nil {
		t.Fatal(err)
	}
	if got := memory.Profile(buyer.ID).CanPurchase; got == nil || !*got {
		t.Fatal("personal account cannot purchase")
	}
	if got := memory.Profile(database.DemoUserID).CanPurchase; got == nil || *got {
		t.Fatal("demo account can purchase")
	}
}

func getLifetimePrice(t *testing.T, handler http.Handler, endpoint string) int {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, endpoint, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", endpoint, response.Code, response.Body.String())
	}
	var result int
	switch endpoint {
	case "/v1/plans":
		var payload struct {
			Data []domain.Plan `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		for _, plan := range payload.Data {
			if plan.BillingType == "lifetime" {
				return plan.PriceIDR
			}
		}
		t.Fatal("lifetime plan missing from plans response")
	case "/v1/home":
		var payload struct {
			Data domain.Dashboard `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		for _, plan := range payload.Data.Plans {
			if plan.BillingType == "lifetime" {
				return plan.PriceIDR
			}
		}
		t.Fatal("lifetime plan missing from home response")
	default:
		var payload struct {
			Data api.PricingConfig `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		result = payload.Data.LifetimePriceIDR
	}
	return result
}
