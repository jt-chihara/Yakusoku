package broker

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// S3Client defines the interface for S3 operations
type S3Client interface {
	PutObject(ctx context.Context, bucket, key string, data []byte) error
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket, prefix string) ([]string, error)
}

// S3Storage implements Storage interface using S3
type S3Storage struct {
	client S3Client
	bucket string
	prefix string

	// Cache for performance (optional, can be disabled)
	mu            sync.RWMutex
	cache         map[string]*contract.Contract
	versions      map[string][]string
	verifications map[string]bool
	cacheEnabled  bool
}

// NewS3Storage creates a new S3-backed storage
func NewS3Storage(client S3Client, bucket, prefix string) *S3Storage {
	return &S3Storage{
		client:        client,
		bucket:        bucket,
		prefix:        prefix,
		cache:         make(map[string]*contract.Contract),
		versions:      make(map[string][]string),
		verifications: make(map[string]bool),
		cacheEnabled:  true,
	}
}

// contractS3Key generates the S3 key for a contract
func (s *S3Storage) contractS3Key(consumer, provider, version string) string {
	return s.prefix + "contracts/" + consumer + "/" + provider + "/" + version + ".json"
}

// verificationS3Key generates the S3 key for verification results
func (s *S3Storage) verificationS3Key(consumer, provider, version string) string {
	return s.prefix + "verifications/" + consumer + "/" + provider + "/" + version + ".json"
}

// indexS3Key generates the S3 key for the index file
func (s *S3Storage) indexS3Key() string {
	return s.prefix + "index.json"
}

// index represents the index structure stored in S3
type index struct {
	Versions      map[string][]string `json:"versions"`       // pairKey -> sorted versions
	Verifications map[string]bool     `json:"verifications"`  // contractKey -> success
}

// loadIndex loads the index from S3
func (s *S3Storage) loadIndex(ctx context.Context) (*index, error) {
	data, err := s.client.GetObject(ctx, s.bucket, s.indexS3Key())
	if err != nil {
		// Return empty index if not found
		return &index{
			Versions:      make(map[string][]string),
			Verifications: make(map[string]bool),
		}, nil
	}

	var idx index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}

	if idx.Versions == nil {
		idx.Versions = make(map[string][]string)
	}
	if idx.Verifications == nil {
		idx.Verifications = make(map[string]bool)
	}

	return &idx, nil
}

// saveIndex saves the index to S3
func (s *S3Storage) saveIndex(ctx context.Context, idx *index) error {
	data, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	return s.client.PutObject(ctx, s.bucket, s.indexS3Key(), data)
}

// SaveContract saves a contract to S3
func (s *S3Storage) SaveContract(c *contract.Contract) error {
	ctx := context.Background()

	// Serialize contract to JSON
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	version := c.Metadata.PactSpecification.Version
	key := s.contractS3Key(c.Consumer.Name, c.Provider.Name, version)

	// Save to S3
	if err := s.client.PutObject(ctx, s.bucket, key, data); err != nil {
		return err
	}

	// Update index
	idx, err := s.loadIndex(ctx)
	if err != nil {
		return err
	}

	pk := pairKey(c.Consumer.Name, c.Provider.Name)
	versions := idx.Versions[pk]
	found := false
	for _, v := range versions {
		if v == version {
			found = true
			break
		}
	}
	if !found {
		versions = append(versions, version)
		sort.Strings(versions)
		idx.Versions[pk] = versions
	}

	if err := s.saveIndex(ctx, idx); err != nil {
		return err
	}

	// Update cache
	if s.cacheEnabled {
		s.mu.Lock()
		s.cache[contractKey(c.Consumer.Name, c.Provider.Name, version)] = c
		s.versions[pk] = versions
		s.mu.Unlock()
	}

	return nil
}

// GetContract retrieves a contract from S3
func (s *S3Storage) GetContract(consumer, provider, version string) (*contract.Contract, error) {
	ctx := context.Background()

	// If version is empty, get the latest
	if version == "" {
		idx, err := s.loadIndex(ctx)
		if err != nil {
			return nil, err
		}

		pk := pairKey(consumer, provider)
		versions := idx.Versions[pk]
		if len(versions) == 0 {
			return nil, ErrNotFound
		}
		version = versions[len(versions)-1]
	}

	// Check cache first
	if s.cacheEnabled {
		s.mu.RLock()
		if c, ok := s.cache[contractKey(consumer, provider, version)]; ok {
			s.mu.RUnlock()
			return c, nil
		}
		s.mu.RUnlock()
	}

	// Load from S3
	key := s.contractS3Key(consumer, provider, version)
	data, err := s.client.GetObject(ctx, s.bucket, key)
	if err != nil {
		return nil, ErrNotFound
	}

	var c contract.Contract
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	// Update cache
	if s.cacheEnabled {
		s.mu.Lock()
		s.cache[contractKey(consumer, provider, version)] = &c
		s.mu.Unlock()
	}

	return &c, nil
}

// ListContracts returns all contracts (latest version of each pair)
func (s *S3Storage) ListContracts() []contract.Contract {
	ctx := context.Background()

	idx, err := s.loadIndex(ctx)
	if err != nil {
		return []contract.Contract{}
	}

	result := make([]contract.Contract, 0)
	for pk, versions := range idx.Versions {
		if len(versions) == 0 {
			continue
		}

		// Parse consumer and provider from pk
		parts := strings.Split(pk, "|")
		if len(parts) != 2 {
			continue
		}
		consumer, provider := parts[0], parts[1]

		latestVersion := versions[len(versions)-1]
		c, err := s.GetContract(consumer, provider, latestVersion)
		if err == nil {
			result = append(result, *c)
		}
	}

	return result
}

// GetContractsByProvider returns contracts for a specific provider
func (s *S3Storage) GetContractsByProvider(provider string) []contract.Contract {
	ctx := context.Background()

	idx, err := s.loadIndex(ctx)
	if err != nil {
		return []contract.Contract{}
	}

	result := make([]contract.Contract, 0)
	for pk, versions := range idx.Versions {
		if len(versions) == 0 {
			continue
		}

		parts := strings.Split(pk, "|")
		if len(parts) != 2 || parts[1] != provider {
			continue
		}
		consumer := parts[0]

		latestVersion := versions[len(versions)-1]
		c, err := s.GetContract(consumer, provider, latestVersion)
		if err == nil {
			result = append(result, *c)
		}
	}

	return result
}

// GetContractsByConsumer returns contracts for a specific consumer
func (s *S3Storage) GetContractsByConsumer(consumer string) []contract.Contract {
	ctx := context.Background()

	idx, err := s.loadIndex(ctx)
	if err != nil {
		return []contract.Contract{}
	}

	result := make([]contract.Contract, 0)
	for pk, versions := range idx.Versions {
		if len(versions) == 0 {
			continue
		}

		parts := strings.Split(pk, "|")
		if len(parts) != 2 || parts[0] != consumer {
			continue
		}
		provider := parts[1]

		latestVersion := versions[len(versions)-1]
		c, err := s.GetContract(consumer, provider, latestVersion)
		if err == nil {
			result = append(result, *c)
		}
	}

	return result
}

// DeleteContract deletes a contract from S3
func (s *S3Storage) DeleteContract(consumer, provider, version string) error {
	ctx := context.Background()

	key := s.contractS3Key(consumer, provider, version)

	// Check if exists first
	if _, err := s.client.GetObject(ctx, s.bucket, key); err != nil {
		return ErrNotFound
	}

	// Delete from S3
	if err := s.client.DeleteObject(ctx, s.bucket, key); err != nil {
		return err
	}

	// Update index
	idx, err := s.loadIndex(ctx)
	if err != nil {
		return err
	}

	pk := pairKey(consumer, provider)
	versions := idx.Versions[pk]
	newVersions := make([]string, 0, len(versions))
	for _, v := range versions {
		if v != version {
			newVersions = append(newVersions, v)
		}
	}
	idx.Versions[pk] = newVersions

	// Also delete verification if exists
	delete(idx.Verifications, contractKey(consumer, provider, version))

	if err := s.saveIndex(ctx, idx); err != nil {
		return err
	}

	// Update cache
	if s.cacheEnabled {
		s.mu.Lock()
		delete(s.cache, contractKey(consumer, provider, version))
		s.versions[pk] = newVersions
		delete(s.verifications, contractKey(consumer, provider, version))
		s.mu.Unlock()
	}

	return nil
}

// RecordVerification records a verification result
func (s *S3Storage) RecordVerification(consumer, provider, version string, success bool) error {
	ctx := context.Background()

	// Update index
	idx, err := s.loadIndex(ctx)
	if err != nil {
		return err
	}

	idx.Verifications[contractKey(consumer, provider, version)] = success

	if err := s.saveIndex(ctx, idx); err != nil {
		return err
	}

	// Update cache
	if s.cacheEnabled {
		s.mu.Lock()
		s.verifications[contractKey(consumer, provider, version)] = success
		s.mu.Unlock()
	}

	return nil
}

// GetVerification gets verification status
func (s *S3Storage) GetVerification(consumer, provider, version string) (success, exists bool) {
	ctx := context.Background()

	// Check cache first
	if s.cacheEnabled {
		s.mu.RLock()
		if v, ok := s.verifications[contractKey(consumer, provider, version)]; ok {
			s.mu.RUnlock()
			return v, true
		}
		s.mu.RUnlock()
	}

	// Load from index
	idx, err := s.loadIndex(ctx)
	if err != nil {
		return false, false
	}

	success, exists = idx.Verifications[contractKey(consumer, provider, version)]
	return success, exists
}

// IsDeployable checks if a pacticipant version can be deployed
func (s *S3Storage) IsDeployable(pacticipant, version string) (deployable bool, reason string) {
	contracts := s.ListContracts()

	for _, c := range contracts {
		if c.Consumer.Name == pacticipant && c.Metadata.PactSpecification.Version == version {
			success, exists := s.GetVerification(c.Consumer.Name, c.Provider.Name, version)
			if !exists {
				return false, "No verification results found"
			}
			if !success {
				return false, "Verification failed"
			}
		}
	}

	return true, "All required verification results are published and successful"
}
