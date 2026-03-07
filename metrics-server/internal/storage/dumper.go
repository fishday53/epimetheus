package storage

import (
	"metrics-server/internal/usecase/context"
	"time"
)

func Dumper(app *context.AppContext) {
	for {
		app.DB.Dump(app.Cfg.FileStoragePath)
		time.Sleep(time.Duration(app.Cfg.StoreInterval) * time.Second)
	}
}
