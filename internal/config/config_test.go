package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	want := &Config{
		ClientID:     "test-client-id",
		TodoListID:   "abc-123",
		TodoListName: "My Tasks",
	}

	if err := Save(want, dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.ClientID != want.ClientID || got.TodoListID != want.TodoListID || got.TodoListName != want.TodoListName {
		t.Errorf("Load returned %+v, want %+v", got, want)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.TodoListID != "" || cfg.TodoListName != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	base := t.TempDir()
	// Point at a subdirectory that does not exist yet.
	dir := filepath.Join(base, "nested")

	cfg := &Config{TodoListID: "x", TodoListName: "y"}
	if err := Save(cfg, dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	p, err := Path(dir)
	if err != nil {
		t.Fatalf("Path: %v", err)
	}

	if _, err := os.Stat(p); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid",
			cfg:     Config{TodoListID: "id", TodoListName: "name"},
			wantErr: false,
		},
		{
			name:    "missing id",
			cfg:     Config{TodoListName: "name"},
			wantErr: true,
		},
		{
			name:    "missing name",
			cfg:     Config{TodoListID: "id"},
			wantErr: true,
		},
		{
			name:    "both missing",
			cfg:     Config{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
