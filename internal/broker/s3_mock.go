package broker

import (
	"context"
	"errors"
	"sync"
)

// MockS3Client is an in-memory mock implementation of S3Client for testing
type MockS3Client struct {
	mu      sync.RWMutex
	objects map[string]map[string][]byte // bucket -> key -> data
}

// NewMockS3Client creates a new mock S3 client
func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		objects: make(map[string]map[string][]byte),
	}
}

// PutObject stores an object in the mock S3
func (m *MockS3Client) PutObject(ctx context.Context, bucket, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.objects[bucket] == nil {
		m.objects[bucket] = make(map[string][]byte)
	}

	// Make a copy of data to avoid mutation issues
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	m.objects[bucket][key] = dataCopy

	return nil
}

// GetObject retrieves an object from the mock S3
func (m *MockS3Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.objects[bucket] == nil {
		return nil, errors.New("bucket not found")
	}

	data, ok := m.objects[bucket][key]
	if !ok {
		return nil, errors.New("key not found")
	}

	// Return a copy to avoid mutation issues
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	return dataCopy, nil
}

// DeleteObject removes an object from the mock S3
func (m *MockS3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.objects[bucket] == nil {
		return nil
	}

	delete(m.objects[bucket], key)
	return nil
}

// ListObjects lists objects with a given prefix
func (m *MockS3Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.objects[bucket] == nil {
		return []string{}, nil
	}

	var keys []string
	for key := range m.objects[bucket] {
		if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}

	return keys, nil
}
