package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Writer writes contract files.
type Writer struct{}

// NewWriter creates a new Writer.
func NewWriter() *Writer {
	return &Writer{}
}

// Write writes a contract to a file.
func (w *Writer) Write(c Contract, path string) error {
	data, err := w.WriteBytes(c)
	if err != nil {
		return err
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write contract file: %w", err)
	}

	return nil
}

// WriteToDir writes a contract to a directory with auto-generated filename.
func (w *Writer) WriteToDir(c Contract, dir string) (string, error) {
	filename := w.generateFilename(c.Consumer.Name, c.Provider.Name)
	path := filepath.Join(dir, filename)

	if err := w.Write(c, path); err != nil {
		return "", err
	}

	return path, nil
}

// WriteBytes returns a contract as bytes.
func (w *Writer) WriteBytes(c Contract) ([]byte, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contract: %w", err)
	}
	return data, nil
}

func (w *Writer) generateFilename(consumer, provider string) string {
	consumer = w.sanitizeName(consumer)
	provider = w.sanitizeName(provider)
	return fmt.Sprintf("%s-%s.json", consumer, provider)
}

func (w *Writer) sanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
