package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"nihongo-master-backend/internal/domain"
)

type MangaAccessChecker func(context.Context, string) (domain.Entitlement, error)

var errMangaSession = errors.New("invalid manga session")

// The isolated service delegates session and store-purchase validation to the
// account API. It never trusts client entitlement flags or caches access grants.
func NewMangaBillingChecker(baseURL string) MangaAccessChecker {
	client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return func(ctx context.Context, token string) (domain.Entitlement, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(baseURL, "/")+"/billing", nil)
		if err != nil {
			return domain.Entitlement{}, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := client.Do(req)
		if err != nil {
			return domain.Entitlement{}, err
		}
		defer res.Body.Close()
		if res.StatusCode == 401 || res.StatusCode == 403 {
			return domain.Entitlement{}, errMangaSession
		}
		if res.StatusCode != 200 {
			return domain.Entitlement{}, errors.New("billing validation unavailable")
		}
		var data struct {
			Data struct {
				Entitlement domain.Entitlement `json:"entitlement"`
			} `json:"data"`
		}
		if err := json.NewDecoder(io.LimitReader(res.Body, 64*1024)).Decode(&data); err != nil {
			return domain.Entitlement{}, err
		}
		return data.Data.Entitlement, nil
	}
}

func (s *Server) requireMangaLifetime(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		token := bearerToken(r)
		if token == "" {
			writeError(w, 401, "unauthorized", "Masuk untuk mengakses Manga Jepang.")
			return
		}
		var entitlement domain.Entitlement
		if s.store != nil {
			id, ok := s.store.ResolveSession(token)
			if !ok {
				writeError(w, 401, "unauthorized", "Sesi login tidak valid atau telah dicabut.")
				return
			}
			if s.billing != nil {
				if err := s.billing.RefreshUser(r.Context(), id); err != nil {
					writeBillingError(w, err)
					return
				}
			}
			entitlement = s.store.Profile(id).Entitlement
		} else {
			if s.mangaAccess == nil {
				writeError(w, 503, "manga_access_unavailable", "Verifikasi Lifetime Access belum tersedia.")
				return
			}
			var err error
			entitlement, err = s.mangaAccess(r.Context(), token)
			if errors.Is(err, errMangaSession) {
				writeError(w, 401, "unauthorized", "Sesi login tidak valid atau telah dicabut.")
				return
			}
			if err != nil {
				writeError(w, 503, "manga_access_unavailable", "Lifetime Access belum dapat diverifikasi. Coba lagi.")
				return
			}
		}
		if entitlement.Tier != domain.TierPremium || !entitlement.IsLifetime {
			writeError(w, 402, "premium_required", "Manga Jepang khusus pengguna Lifetime Access.")
			return
		}
		next(w, r)
	}
}
