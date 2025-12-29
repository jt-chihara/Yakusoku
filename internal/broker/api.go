package broker

import (
	"encoding/json"
	"net/http"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// API is the HTTP API for the broker
type API struct {
	storage *MemoryStorage
}

// NewAPI creates a new broker API
func NewAPI(storage *MemoryStorage) *API {
	return &API{storage: storage}
}

// Handler returns the HTTP handler for the API
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	// List all contracts
	mux.HandleFunc("GET /pacts", a.handleListContracts)

	// Get contracts by provider
	mux.HandleFunc("GET /pacts/provider/{provider}", a.handleGetContractsByProvider)

	// Get specific contract (with version)
	mux.HandleFunc("GET /pacts/provider/{provider}/consumer/{consumer}/version/{version}", a.handleGetContract)

	// Get latest contract
	mux.HandleFunc("GET /pacts/provider/{provider}/consumer/{consumer}/latest", a.handleGetLatestContract)

	// Publish contract
	mux.HandleFunc("POST /pacts/provider/{provider}/consumer/{consumer}/version/{version}", a.handlePublishContract)
	mux.HandleFunc("PUT /pacts/provider/{provider}/consumer/{consumer}/version/{version}", a.handlePublishContract)

	// Delete contract
	mux.HandleFunc("DELETE /pacts/provider/{provider}/consumer/{consumer}/version/{version}", a.handleDeleteContract)

	// Record verification result
	mux.HandleFunc("POST /pacts/provider/{provider}/consumer/{consumer}/version/{version}/verification-results", a.handleRecordVerification)

	// Matrix / Can I Deploy
	mux.HandleFunc("GET /matrix", a.handleMatrix)

	return mux
}

func (a *API) handleListContracts(w http.ResponseWriter, r *http.Request) {
	contracts := a.storage.ListContracts()

	result := make([]map[string]interface{}, len(contracts))
	for i, c := range contracts {
		result[i] = map[string]interface{}{
			"consumer": c.Consumer.Name,
			"provider": c.Provider.Name,
			"version":  c.Metadata.PactSpecification.Version,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (a *API) handleGetContractsByProvider(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")

	contracts := a.storage.GetContractsByProvider(provider)

	result := make([]map[string]interface{}, len(contracts))
	for i, c := range contracts {
		result[i] = map[string]interface{}{
			"consumer": c.Consumer.Name,
			"provider": c.Provider.Name,
			"version":  c.Metadata.PactSpecification.Version,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (a *API) handleGetContract(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	consumer := r.PathValue("consumer")
	version := r.PathValue("version")

	c, err := a.storage.GetContract(consumer, provider, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

func (a *API) handleGetLatestContract(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	consumer := r.PathValue("consumer")

	c, err := a.storage.GetContract(consumer, provider, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

func (a *API) handlePublishContract(w http.ResponseWriter, r *http.Request) {
	var c contract.Contract
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Override consumer/provider/version from URL
	c.Consumer.Name = r.PathValue("consumer")
	c.Provider.Name = r.PathValue("provider")
	c.Metadata.PactSpecification.Version = r.PathValue("version")

	if err := a.storage.SaveContract(&c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (a *API) handleDeleteContract(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	consumer := r.PathValue("consumer")
	version := r.PathValue("version")

	if err := a.storage.DeleteContract(consumer, provider, version); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleRecordVerification(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	consumer := r.PathValue("consumer")
	version := r.PathValue("version")

	var body struct {
		Success         bool   `json:"success"`
		ProviderVersion string `json:"providerVersion"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := a.storage.RecordVerification(consumer, provider, version, body.Success); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

func (a *API) handleMatrix(w http.ResponseWriter, r *http.Request) {
	pacticipant := r.URL.Query().Get("pacticipant")
	version := r.URL.Query().Get("version")

	deployable, reason := a.storage.IsDeployable(pacticipant, version)

	result := map[string]interface{}{
		"deployable": deployable,
		"summary": map[string]interface{}{
			"deployable": deployable,
			"reason":     reason,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
