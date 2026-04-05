// This part of storage package implements periodic dump for metric storage.
package storage

import (
	"metrics-server/internal/usecase/context"
	"time"
)

// Dumper implements periodic dump for metric storage.
func Dumper(app *context.AppContext) {
	for {
		app.DB.Dump(app.Cfg.FileStoragePath)
		time.Sleep(time.Duration(app.Cfg.StoreInterval) * time.Second)
	}
}
