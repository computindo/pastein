package main

import (
		"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"pastein/cleanup"
	"pastein/handler"
	"pastein/logger"
	"pastein/storage"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Printf("WARNING: env var %q invalid integer value %q, using default", key, v)
			return def
		}
		return i
	}
	return def
}

func main() {
	dataDir := getEnv("DATA_DIR", "./data")
	baseURL := getEnv("BASE_URL", "http://localhost:8282")
	maxUpload := getEnvInt("MAX_UPLOAD_SIZE", 52428800)
	defaultExpiry := getEnvInt("DEFAULT_EXPIRY", 86400)
	cleanupInterval := getEnvInt("CLEANUP_INTERVAL", 900)
	port := getEnv("PORT", "8282")

	store := storage.New(dataDir)
	go cleanup.Start(store, time.Duration(cleanupInterval)*time.Second)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t := time.Now()
			next.ServeHTTP(ww, r)
			logger.Info("Request", fmt.Sprintf("%s %s — %d %dB %v",
				r.Method, r.URL.Path, ww.Status(), ww.BytesWritten(), time.Since(t).Round(time.Microsecond)))
		})
	})
	r.Use(middleware.Recoverer)

	h := handler.New(store, baseURL, maxUpload, defaultExpiry)
	r.Get("/", h.ServeIndex)
	r.Get("/{slug}", h.ServeView)
	r.Post("/api/snippets", h.CreateSnippet)
	r.Post("/api/files", h.UploadFile)
	r.Get("/recent", h.GetRecentPage)
	r.Get("/api/recent", h.GetRecentAPI)
	r.Get("/api/snippets/{slug}", h.GetSnippet)
	r.Get("/files/{slug}/{filename}", h.ServeFile)

	staticDir := getEnv("STATIC_DIR", "./static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	addr := ":" + port
	logger.Info("Server", "PASTEIN starting on "+addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("Server", "Fatal: "+err.Error())
	}
}
