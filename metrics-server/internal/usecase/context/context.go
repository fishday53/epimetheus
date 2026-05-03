// Package context implements transparent set of Metrics-Server settings for server-wide usage.
package context

import (
	"crypto/rsa"
	"fmt"
	"metrics-server/internal/config"
	"metrics-server/internal/crypt"
	"metrics-server/internal/log"
	"metrics-server/internal/storage/memory"
	"metrics-server/internal/storage/postgres"
	"metrics-server/internal/usecase"

	"go.uber.org/zap"
)

// AppContext stores server-wide settings.
type AppContext struct {
	DB      usecase.Repository
	Log     *zap.SugaredLogger
	Cfg     *config.Config
	PrivKey *rsa.PrivateKey
	Stop    chan struct{}
}

// NewAppContext initializes a new AppContext.
func NewAppContext(cfg *config.Config) (*AppContext, error) {
	var err error
	a := AppContext{
		Log: log.NewLogger(),
		Cfg: cfg,
	}

	if cfg.DSN == "" {
		a.DB = memory.NewMemStorage()
	} else {
		a.DB, err = postgres.NewPsqlStorage(cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("cannot initialize new app context: %v", err)
		}
	}

	if cfg.CryptoKeyPath != "" {
		a.PrivKey, err = crypt.GetPrivateKey(cfg.CryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("cannot get %s: %v", cfg.CryptoKeyPath, err)
		}
	}

	a.Stop = make(chan struct{})

	return &a, nil
}
