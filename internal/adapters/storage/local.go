package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/logger"
	"github.com/dariomba/screen-go/internal/ports"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if basePath == "" {
		return nil, errors.New("base path cannot be empty")
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Ctx(ctx).Debug().Msg("File not found")

			return nil, domain.ErrScreenshotNotFound
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (s *LocalStorage) Save(ctx context.Context, input *ports.SaveScreenshotInput) (*ports.SaveScreenshotResult, error) {
	fullPath := filepath.Join(s.basePath, input.Key)

	logger.Ctx(ctx).Debug().
		Str("path", fullPath).
		Msg("Writing to local filesystem")

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	_, err = io.Copy(file, input.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	return &ports.SaveScreenshotResult{Key: input.Key, Size: stat.Size()}, nil
}
