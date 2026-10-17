package auth

import (
	"bytes"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPhotoFilesAreStoredUnderRandomSafeNames(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "photos")
	s := &Service{}
	if err := s.SetPhotoDir(dir); err != nil {
		t.Fatalf("SetPhotoDir: %v", err)
	}
	if info, err := os.Stat(dir); err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("photo directory = %v, %v; want mode 0700", info, err)
	}
	data := []byte("not-really-an-image-but-bytes")
	name, err := s.savePhotoFile(data, "image/png")
	if err != nil || !photoFileNamePattern.MatchString(name) || filepath.Ext(name) != ".png" {
		t.Fatalf("savePhotoFile = %q, %v", name, err)
	}
	other, _ := s.savePhotoFile(data, "image/jpeg")
	if other == name || filepath.Ext(other) != ".jpg" {
		t.Fatalf("names must be unique and follow the type: %q vs %q", name, other)
	}
	if info, err := os.Stat(filepath.Join(dir, name)); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("photo file = %v, %v; want mode 0600", info, err)
	}
	got, err := s.photoBytes(sql.NullString{String: name, Valid: true}, nil)
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("photoBytes = %q, %v", got, err)
	}
	s.removePhotoFile(name)
	if _, err := s.photoBytes(sql.NullString{String: name, Valid: true}, nil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("removed photo error = %v, want not-exist", err)
	}
	// Legacy rows keep serving their blob.
	if got, err := s.photoBytes(sql.NullString{}, []byte("blob")); err != nil || string(got) != "blob" {
		t.Fatalf("blob fallback = %q, %v", got, err)
	}
}

func TestPhotoFileNamesCannotEscapeTheDirectory(t *testing.T) {
	s := &Service{}
	if err := s.SetPhotoDir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"../etc/passwd", "..%2f..%2fsecret.png", "a/b.png", "/etc/passwd",
		"0123456789abcdef0123456789abcdef.svg", "0123456789abcdef0123456789abcde.png", "",
	} {
		if _, err := s.photoBytes(sql.NullString{String: name, Valid: true}, nil); !errors.Is(err, errUnsafePhotoName) {
			t.Errorf("photoBytes(%q) error = %v, want errUnsafePhotoName", name, err)
		}
	}
	if _, err := s.savePhotoFile([]byte("x"), "image/svg+xml"); err == nil {
		t.Error("unsupported type accepted")
	}
}

func TestSetPhotoDirRejectsUnwritableDirectory(t *testing.T) {
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(parent, 0o700)
	if os.Geteuid() == 0 {
		t.Skip("root can write anywhere")
	}
	if err := (&Service{}).SetPhotoDir(filepath.Join(parent, "photos")); err == nil {
		t.Fatal("unwritable directory accepted")
	}
}
