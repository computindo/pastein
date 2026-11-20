package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

func (h *Handler) GetSnippet(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	item, err := h.store.Get(slug)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	if !item.Permanent && item.ExpiresAt != "" {
		expires, err := time.Parse(time.RFC3339, item.ExpiresAt)
		if err == nil && time.Now().After(expires) {
			http.Error(w, `{"error":"expired"}`, http.StatusGone)
			return
		}
	}

	resp := map[string]interface{}{
		"slug":       item.Slug,
		"type":       item.Type,
		"created_at": item.CreatedAt,
		"permanent":  item.Permanent,
	}

	if !item.Permanent && item.ExpiresAt != "" {
		resp["expires_at"] = item.ExpiresAt
	}

	if item.Type == "text" {
		resp["content"] = item.Content
		resp["language"] = item.Language
	}

	if item.Type == "file" {
		resp["filename"] = item.Filename
		resp["mime_type"] = item.MimeType
		resp["file_size"] = item.FileSize
		resp["url"] = h.baseURL + "/files/" + item.Slug + "/" + url.PathEscape(item.Filename)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
