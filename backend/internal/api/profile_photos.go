package api

import (
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"nihongo-master-backend/internal/media"
)

const maxProfilePhotoRequestBytes = media.MaxProfilePhotoUploadBytes + (1 << 20)

func (s *Server) uploadProfilePhoto(w http.ResponseWriter, r *http.Request) {
	if s.profilePhotos == nil {
		writeError(w, http.StatusServiceUnavailable, "photo_upload_unavailable", "Upload foto profil belum dikonfigurasi di server.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxProfilePhotoRequestBytes)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) || err == multipart.ErrMessageTooLarge {
			writeError(w, http.StatusRequestEntityTooLarge, "photo_too_large", "Ukuran foto maksimal 8 MB.")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_photo_upload", "Pilih file foto yang valid lalu coba lagi.")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "photo_required", "File foto belum dipilih.")
		return
	}
	defer file.Close()
	input, err := io.ReadAll(io.LimitReader(file, media.MaxProfilePhotoUploadBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_photo_upload", "File foto tidak dapat dibaca.")
		return
	}
	if len(input) > media.MaxProfilePhotoUploadBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "photo_too_large", "Ukuran foto maksimal 8 MB.")
		return
	}
	photoURL, err := s.profilePhotos.Upload(r.Context(), userID(r), input)
	if err != nil {
		if errors.Is(err, media.ErrInvalidProfilePhoto) {
			message := strings.TrimSpace(strings.TrimPrefix(err.Error(), media.ErrInvalidProfilePhoto.Error()+":"))
			writeError(w, http.StatusUnprocessableEntity, "invalid_photo", message)
			return
		}
		writeError(w, http.StatusBadGateway, "photo_upload_failed", "Foto gagal diunggah ke penyimpanan. Coba lagi.")
		return
	}
	profile, previousURL, err := s.store.UpdateProfilePhoto(userID(r), photoURL)
	if err != nil {
		if cleanupErr := s.profilePhotos.Delete(r.Context(), userID(r), photoURL); cleanupErr != nil {
			log.Printf("cleanup uncommitted profile photo: %v", cleanupErr)
		}
		writeStoreError(w, err)
		return
	}
	if previousURL != "" {
		if err := s.profilePhotos.Delete(r.Context(), userID(r), previousURL); err != nil {
			log.Printf("cleanup replaced profile photo: %v", err)
		}
	}
	writeData(w, http.StatusOK, profile)
}
