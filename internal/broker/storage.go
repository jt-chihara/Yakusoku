package broker

import (
	"errors"
	"sort"
	"sync"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// ErrNotFound indicates that the requested contract was not found
var ErrNotFound = errors.New("contract not found")

// Storage defines the interface for contract storage backends
type Storage interface {
	SaveContract(c *contract.Contract) error
	GetContract(consumer, provider, version string) (*contract.Contract, error)
	ListContracts() []contract.Contract
	GetContractsByProvider(provider string) []contract.Contract
	GetContractsByConsumer(consumer string) []contract.Contract
	DeleteContract(consumer, provider, version string) error
	RecordVerification(consumer, provider, version string, success bool) error
	GetVerification(consumer, provider, version string) (success, exists bool)
	IsDeployable(pacticipant, version string) (deployable bool, reason string)
}

// contractKey generates a unique key for a contract
func contractKey(consumer, provider, version string) string {
	return consumer + "|" + provider + "|" + version
}

// pairKey generates a key for a consumer-provider pair
func pairKey(consumer, provider string) string {
	return consumer + "|" + provider
}

// MemoryStorage is an in-memory storage for contracts
type MemoryStorage struct {
	mu            sync.RWMutex
	contracts     map[string]contract.Contract
	versions      map[string][]string // pairKey -> sorted versions
	verifications map[string]bool     // contractKey -> success
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		contracts:     make(map[string]contract.Contract),
		versions:      make(map[string][]string),
		verifications: make(map[string]bool),
	}
}

// SaveContract saves a contract to storage
func (s *MemoryStorage) SaveContract(c *contract.Contract) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	version := c.Metadata.PactSpecification.Version
	key := contractKey(c.Consumer.Name, c.Provider.Name, version)
	pk := pairKey(c.Consumer.Name, c.Provider.Name)

	s.contracts[key] = *c

	// Update version list
	versions := s.versions[pk]
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
		s.versions[pk] = versions
	}

	return nil
}

// GetContract retrieves a contract from storage
func (s *MemoryStorage) GetContract(consumer, provider, version string) (*contract.Contract, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// If version is empty, get the latest
	if version == "" {
		pk := pairKey(consumer, provider)
		versions := s.versions[pk]
		if len(versions) == 0 {
			return nil, ErrNotFound
		}
		version = versions[len(versions)-1]
	}

	key := contractKey(consumer, provider, version)
	c, ok := s.contracts[key]
	if !ok {
		return nil, ErrNotFound
	}

	return &c, nil
}

// ListContracts returns all contracts (latest version of each pair)
func (s *MemoryStorage) ListContracts() []contract.Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]contract.Contract, 0)
	seen := make(map[string]bool)

	for pk, versions := range s.versions {
		if seen[pk] {
			continue
		}
		seen[pk] = true

		if len(versions) > 0 {
			latestVersion := versions[len(versions)-1]
			// Parse consumer and provider from pk
			for key, c := range s.contracts {
				cpk := pairKey(c.Consumer.Name, c.Provider.Name)
				if cpk == pk && c.Metadata.PactSpecification.Version == latestVersion {
					result = append(result, s.contracts[key])
					break
				}
			}
		}
	}

	return result
}

// GetContractsByProvider returns contracts for a specific provider
func (s *MemoryStorage) GetContractsByProvider(provider string) []contract.Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]contract.Contract, 0)
	seen := make(map[string]bool)

	for _, c := range s.contracts {
		if c.Provider.Name != provider {
			continue
		}
		pk := pairKey(c.Consumer.Name, c.Provider.Name)
		if seen[pk] {
			continue
		}
		seen[pk] = true

		// Get latest version for this pair
		versions := s.versions[pk]
		if len(versions) > 0 {
			latestVersion := versions[len(versions)-1]
			key := contractKey(c.Consumer.Name, c.Provider.Name, latestVersion)
			if latest, ok := s.contracts[key]; ok {
				result = append(result, latest)
			}
		}
	}

	return result
}

// GetContractsByConsumer returns contracts for a specific consumer
func (s *MemoryStorage) GetContractsByConsumer(consumer string) []contract.Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]contract.Contract, 0)
	seen := make(map[string]bool)

	for _, c := range s.contracts {
		if c.Consumer.Name != consumer {
			continue
		}
		pk := pairKey(c.Consumer.Name, c.Provider.Name)
		if seen[pk] {
			continue
		}
		seen[pk] = true

		// Get latest version for this pair
		versions := s.versions[pk]
		if len(versions) > 0 {
			latestVersion := versions[len(versions)-1]
			key := contractKey(c.Consumer.Name, c.Provider.Name, latestVersion)
			if latest, ok := s.contracts[key]; ok {
				result = append(result, latest)
			}
		}
	}

	return result
}

// DeleteContract deletes a contract from storage
func (s *MemoryStorage) DeleteContract(consumer, provider, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := contractKey(consumer, provider, version)
	if _, ok := s.contracts[key]; !ok {
		return ErrNotFound
	}

	delete(s.contracts, key)
	delete(s.verifications, key)

	// Update version list
	pk := pairKey(consumer, provider)
	versions := s.versions[pk]
	newVersions := make([]string, 0, len(versions))
	for _, v := range versions {
		if v != version {
			newVersions = append(newVersions, v)
		}
	}
	s.versions[pk] = newVersions

	return nil
}

// RecordVerification records a verification result
func (s *MemoryStorage) RecordVerification(consumer, provider, version string, success bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := contractKey(consumer, provider, version)
	s.verifications[key] = success
	return nil
}

// GetVerification gets verification status
// Returns (success, exists)
func (s *MemoryStorage) GetVerification(consumer, provider, version string) (success, exists bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := contractKey(consumer, provider, version)
	success, exists = s.verifications[key]
	return success, exists
}

// IsDeployable checks if a pacticipant version can be deployed
func (s *MemoryStorage) IsDeployable(pacticipant, version string) (deployable bool, reason string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find all contracts involving this pacticipant
	for _, c := range s.contracts {
		if c.Consumer.Name == pacticipant && c.Metadata.PactSpecification.Version == version {
			key := contractKey(c.Consumer.Name, c.Provider.Name, version)
			success, exists := s.verifications[key]
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
