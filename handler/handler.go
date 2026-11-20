package handler

import (
	"os"
	"crypto/rand"
		"encoding/json"
	"net/http"
	"time"

	"pastein/logger"
	"pastein/storage"
)

type Handler struct {
	staticDir    string
	store        *storage.Storage
	baseURL      string
	maxUpload    int
	defaultExpiry int
}

func New(store *storage.Storage, baseURL string, maxUpload, defaultExpiry int) *Handler {
	return &Handler{
		staticDir:    func() string { if v := os.Getenv("STATIC_DIR"); v != "" { return v }; return "./static" }(),
		store:        store,
		baseURL:      baseURL,
		maxUpload:    maxUpload,
		defaultExpiry: defaultExpiry,
	}
}

func generateSlug() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const max = 256 - (256 % len(charset)) // rejection threshold: sampling bebas bias
	slug := make([]byte, 8)
	buf := make([]byte, 16)
	i := 0
	for i < len(slug) {
		_, _ = rand.Read(buf)
		for _, b := range buf {
			if i == len(slug) {
				break
			}
			if int(b) < max {
				slug[i] = charset[int(b)%len(charset)]
				i++
			}
		}
	}
	return string(slug)
}

func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.staticDir+"/index.html")
}

func (h *Handler) ServeView(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.staticDir+"/view.html")
}

func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	
	item, err := h.store.Get(slug)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	
	if item.Type != "file" {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, item.FilePath)
}

func (h *Handler) GetRecentPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.staticDir+"/recent.html")
}

func (h *Handler) GetRecentAPI(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListRecent()
	if err != nil {
		logger.Error("recent", "failed to list recent: "+err.Error())
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

type SnippetRequest struct {
	Content   string `json:"content"`
	Language  string `json:"language"`
	ExpiresIn int    `json:"expires_in"`
}

type SnippetResponse struct {
	Slug      string     `json:"slug"`
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expires_at"`
	Type      string     `json:"type"`
}

func (h *Handler) CreateSnippet(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit
	var req SnippetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Language == "" {
		req.Language = "plaintext"
	}
	permanent := req.ExpiresIn == 0
	if !permanent {
		req.ExpiresIn = h.defaultExpiry
	}

	slug := generateSlug()
	var expiresAt string
	var expiresTime *time.Time
	if !permanent {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = t.Format(time.RFC3339)
		expiresTime = &t
	}

	for i := 0; i < 3; i++ {
		if !h.store.Exists(slug) {
			break
		}
		slug = generateSlug()
		if i == 2 {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
	}

	now := time.Now().Format(time.RFC3339)
	item := &storage.Item{
		Slug:      slug,
		Type:      "text",
		Content:   req.Content,
		Language:  req.Language,
		Permanent: permanent,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	if err := h.store.Save(item); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	resp := SnippetResponse{
		Slug:      slug,
		URL:       h.baseURL + "/" + slug,
		ExpiresAt: expiresTime,
		Type:      "text",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
