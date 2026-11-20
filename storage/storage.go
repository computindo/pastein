package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Item struct {
	Slug      string `json:"slug"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Language  string `json:"language"`
	Filename  string `json:"filename"`
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	Permanent bool   `json:"permanent"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
}

type Storage struct {
	dataDir    string
	itemsDir   string
	uploadsDir string
}

func New(dataDir string) *Storage {
	s := &Storage{
		dataDir:    dataDir,
		itemsDir:   filepath.Join(dataDir, "items"),
		uploadsDir: filepath.Join(dataDir, "uploads"),
	}
	os.MkdirAll(s.itemsDir, 0755)
	os.MkdirAll(s.uploadsDir, 0755)
	return s
}

func (s *Storage) Save(item *Item) error {
	path := filepath.Join(s.itemsDir, item.Slug+".json")
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) Get(slug string) (*Item, error) {
	path := filepath.Join(s.itemsDir, slug+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var item Item
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Storage) Exists(slug string) bool {
	path := filepath.Join(s.itemsDir, slug+".json")
	_, err := os.Stat(path)
	return err == nil
}

func (s *Storage) Delete(slug string) error {
	item, err := s.Get(slug)
	if err != nil {
		return err
	}
	if item.FilePath != "" {
		os.Remove(item.FilePath)
	}
	path := filepath.Join(s.itemsDir, slug+".json")
	return os.Remove(path)
}

func (s *Storage) ListExpired() ([]*Item, error) {
	files, err := os.ReadDir(s.itemsDir)
	if err != nil {
		return nil, err
	}
	var items []*Item
	now := time.Now()
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		slug := f.Name()[:len(f.Name())-5]
		item, err := s.Get(slug)
		if err != nil {
			continue
		}
		if !item.Permanent && item.ExpiresAt != "" {
			expires, err := time.Parse(time.RFC3339, item.ExpiresAt)
			if err == nil && now.After(expires) {
				items = append(items, item)
			}
		}
	}
	return items, nil
}

func (s *Storage) UploadsDir() string {
	return s.uploadsDir
}

func (s *Storage) DataDir() string {
	return s.dataDir
}

// ListRecent mengembalikan semua item yang belum expired, terbaru dulu.
func (s *Storage) ListRecent() ([]*Item, error) {
	entries, err := os.ReadDir(s.itemsDir)
	if err != nil {
		return nil, err
	}
	items := []*Item{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		item, err := s.Get(e.Name()[:len(e.Name())-5])
		if err != nil {
			continue
		}
		if !item.Permanent && item.ExpiresAt != "" {
			if expires, err := time.Parse(time.RFC3339, item.ExpiresAt); err == nil && time.Now().After(expires) {
				continue
			}
		}
		items = append(items, item)
	}
	// RFC3339 UTC dari server: urutan leksikografis = urutan waktu
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items, nil
}
