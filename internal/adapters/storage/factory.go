package storage

import (
	"fmt"

	"github.com/dariomba/screen-go/internal/ports"
)

type StorageProvider string

const (
	ProviderFilesystem StorageProvider = "filesystem"
	ProviderS3         StorageProvider = "s3"
)

type Config struct {
	Provider StorageProvider

	BasePath string

	Bucket    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
}

func NewScreenshotStorage(cfg Config) (ports.ScreenshotStorage, error) {
	switch cfg.Provider {
	case ProviderFilesystem:
		return NewLocalStorage(cfg.BasePath)

	case ProviderS3:
		return NewS3Storage(S3Config{
			Bucket:    cfg.Bucket,
			Endpoint:  cfg.Endpoint,
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
			Region:    cfg.Region,
		})

	default:
		return nil, fmt.Errorf("unknown storage provider: %s", cfg.Provider)
	}
}
