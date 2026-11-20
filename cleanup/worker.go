package cleanup

import (
	"fmt"
	"time"

	"pastein/logger"
	"pastein/storage"
)

func Start(store *storage.Storage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cleanup(store)
	}
}

func cleanup(store *storage.Storage) {
	items, err := store.ListExpired()
	if err != nil {
		logger.Error("cleanup", "error listing expired items: "+err.Error())
		return
	}

	for _, item := range items {
		if err := store.Delete(item.Slug); err != nil {
			logger.Error("cleanup", "error deleting "+item.Slug+": "+err.Error())
		} else {
			logger.Info("cleanup", "deleted expired item "+item.Slug)
		}
	}

	if len(items) > 0 {
		logger.Info("cleanup", fmt.Sprintf("removed %d expired items", len(items)))
	}
}
