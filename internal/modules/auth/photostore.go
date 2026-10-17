package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Profile photos are stored as files in a persistent directory, with only a
// server-generated file name in MySQL. Without a configured directory the
// photo falls back to the legacy MEDIUMBLOB column.
var photoFileNamePattern = regexp.MustCompile(`^[0-9a-f]{32}\.(jpg|png)$`)

var errUnsafePhotoName = errors.New("unsafe photo file name")

// SetPhotoDir enables file storage in dir, creating it if needed and checking
// that it is writable.
func (s *Service) SetPhotoDir(dir string) error {
	if dir == "" {
		s.photoDir = ""
		return nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create photo directory: %w", err)
	}
	probe, err := os.CreateTemp(dir, ".write-check-*")
	if err != nil {
		return fmt.Errorf("photo directory is not writable: %w", err)
	}
	probe.Close()
	_ = os.Remove(probe.Name())
	s.photoDir = dir
	return nil
}

func photoExtension(mimeType string) (string, error) {
	switch mimeType {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	}
	return "", ErrInvalidProfilePhoto
}

func (s *Service) photoPath(name string) (string, error) {
	if !photoFileNamePattern.MatchString(name) {
		return "", errUnsafePhotoName
	}
	return filepath.Join(s.photoDir, name), nil
}

// savePhotoFile writes the already-sanitized image under a random name.
func (s *Service) savePhotoFile(data []byte, mimeType string) (string, error) {
	extension, err := photoExtension(mimeType)
	if err != nil {
		return "", err
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate photo file name: %w", err)
	}
	name := hex.EncodeToString(random) + extension
	path, err := s.photoPath(name)
	if err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("create photo file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write photo file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close photo file: %w", err)
	}
	return name, nil
}

func (s *Service) removePhotoFile(name string) {
	if path, err := s.photoPath(name); err == nil {
		_ = os.Remove(path)
	}
}

// storePhoto records the member's photo inside tx. The returned function
// deletes a newly written file and must be called if the transaction does not
// commit; it is a no-op in blob mode.
func (s *Service) storePhoto(ctx context.Context, tx *sql.Tx, accountID, data []byte, mimeType string) (func(), error) {
	if s.photoDir == "" {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO alumni_profile_photos (account_id, mime_type, image_data)
			VALUES (?, ?, ?)`, accountID, mimeType, data)
		return func() {}, err
	}
	name, err := s.savePhotoFile(data, mimeType)
	if err != nil {
		return func() {}, err
	}
	discard := func() { s.removePhotoFile(name) }
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO alumni_profile_photos (account_id, mime_type, file_name)
		VALUES (?, ?, ?)`, accountID, mimeType, name); err != nil {
		discard()
		return func() {}, err
	}
	return discard, nil
}

// photoBytes returns the image for a stored row, from disk when it has a file
// name and from the legacy blob otherwise.
func (s *Service) photoBytes(fileName sql.NullString, blob []byte) ([]byte, error) {
	if !fileName.Valid {
		return blob, nil
	}
	path, err := s.photoPath(fileName.String)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// MigratePhotosToFiles moves photos still stored as blobs onto disk. It is
// idempotent and does nothing unless a photo directory is configured.
func (s *Service) MigratePhotosToFiles(ctx context.Context) (int, error) {
	if s.photoDir == "" {
		return 0, nil
	}
	moved := 0
	for {
		rows, err := s.db.QueryContext(ctx, `
			SELECT account_id, mime_type, image_data FROM alumni_profile_photos
			WHERE file_name IS NULL AND image_data IS NOT NULL LIMIT 10`)
		if err != nil {
			return moved, fmt.Errorf("list legacy photos: %w", err)
		}
		type legacy struct {
			account  []byte
			mimeType string
			data     []byte
		}
		var batch []legacy
		for rows.Next() {
			var item legacy
			if err := rows.Scan(&item.account, &item.mimeType, &item.data); err != nil {
				rows.Close()
				return moved, fmt.Errorf("read legacy photo: %w", err)
			}
			batch = append(batch, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return moved, fmt.Errorf("iterate legacy photos: %w", err)
		}
		rows.Close()
		if len(batch) == 0 {
			return moved, nil
		}
		for _, item := range batch {
			name, err := s.savePhotoFile(item.data, item.mimeType)
			if err != nil {
				return moved, err
			}
			result, err := s.db.ExecContext(ctx, `
				UPDATE alumni_profile_photos SET file_name = ?, image_data = NULL
				WHERE account_id = ? AND file_name IS NULL`, name, item.account)
			if err != nil {
				s.removePhotoFile(name)
				return moved, fmt.Errorf("record migrated photo: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected == 0 {
				s.removePhotoFile(name)
				continue
			}
			moved++
		}
	}
}
