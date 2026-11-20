package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"pastein/storage"
)

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(h.maxUpload))
	if err := r.ParseMultipartForm(8 << 20); err != nil { // 8MB memory buffer, sisanya spill ke temp file
		http.Error(w, `{"error":"content too large"}`, http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	expiresInStr := r.FormValue("expires_in")
	expiresIn := h.defaultExpiry
	if expiresInStr != "" {
		expiresIn, _ = strconv.Atoi(expiresInStr)
	}

	slug := generateSlug()
	permanent := expiresIn == 0
	var expiresAt string
	var expiresTime *time.Time
	if !permanent {
		t := time.Now().Add(time.Duration(expiresIn) * time.Second)
		expiresAt = t.Format(time.RFC3339)
		expiresTime = &t
	}

	os.MkdirAll(h.store.UploadsDir(), 0755)
	filename := fmt.Sprintf("%s_%s", slug, filepath.Base(header.Filename))
	filePath := filepath.Join(h.store.UploadsDir(), filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	for i := 0; i < 3; i++ {
		if !h.store.Exists(slug) {
			break
		}
		slug = generateSlug()
		filename = fmt.Sprintf("%s_%s", slug, filepath.Base(header.Filename))
		filePath = filepath.Join(h.store.UploadsDir(), filename)
		if i == 2 {
			os.Remove(filePath)
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	now := time.Now().Format(time.RFC3339)
	item := &storage.Item{
		Slug:      slug,
		Type:      "file",
		Filename:  header.Filename,
		FilePath:  filePath,
		FileSize:  size,
		MimeType:  header.Header.Get("Content-Type"),
		Permanent: permanent,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	if err := h.store.Save(item); err != nil {
		os.Remove(filePath)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	resp := SnippetResponse{
		Slug:      slug,
		URL:       h.baseURL + "/" + slug,
		ExpiresAt: expiresTime,
		Type:      "file",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
