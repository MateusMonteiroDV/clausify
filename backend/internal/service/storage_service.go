package service

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// StorageService handles local disk file storage.
type StorageService struct {
	baseDir string // e.g. "uploads"
}

func NewStorageService(baseDir string) *StorageService {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create uploads dir: %v", err))
	}
	return &StorageService{baseDir: baseDir}
}

// Save persists the multipart file to disk and returns (storagePath, sha256Hash, size, error).
func (s *StorageService) Save(orgID uuid.UUID, fh *multipart.FileHeader) (storagePath, hash string, size int64, err error) {
	src, err := fh.Open()
	if err != nil {
		return "", "", 0, fmt.Errorf("open upload: %w", err)
	}
	defer src.Close()

	// Organise by org to keep tenants isolated
	dir := filepath.Join(s.baseDir, orgID.String())
	if err = os.MkdirAll(dir, 0755); err != nil {
		return "", "", 0, fmt.Errorf("mkdir: %w", err)
	}

	// Unique filename to avoid collisions
	ext := filepath.Ext(fh.Filename)
	fileName := uuid.New().String() + ext
	destPath := filepath.Join(dir, fileName)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", "", 0, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)

	size, err = io.Copy(writer, src)
	if err != nil {
		return "", "", 0, fmt.Errorf("write file: %w", err)
	}

	hash = fmt.Sprintf("%x", hasher.Sum(nil))
	storagePath = destPath
	return storagePath, hash, size, nil
}

// AbsPath returns the absolute path for a stored file.
func (s *StorageService) AbsPath(storagePath string) string {
	if filepath.IsAbs(storagePath) {
		return storagePath
	}
	abs, _ := filepath.Abs(storagePath)
	return abs
}

// Delete removes a file from disk.
func (s *StorageService) Delete(storagePath string) error {
	return os.Remove(storagePath)
}
