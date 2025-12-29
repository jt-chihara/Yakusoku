package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

func TestContract_JSONMarshaling(t *testing.T) {
	t.Run("marshal contract to JSON", func(t *testing.T) {
		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "OrderService"},
			Provider: contract.Pacticipant{Name: "UserService"},
			Interactions: []contract.Interaction{
				{
					Description:   "a request for user 1",
					ProviderState: "user 1 exists",
					Request: contract.Request{
						Method: "GET",
						Path:   "/users/1",
					},
					Response: contract.Response{
						Status: 200,
						Body:   map[string]interface{}{"id": float64(1), "name": "John Doe"},
					},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		data, err := json.Marshal(c)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "OrderService", result["consumer"].(map[string]interface{})["name"])
		assert.Equal(t, "UserService", result["provider"].(map[string]interface{})["name"])
		assert.Len(t, result["interactions"], 1)
	})

	t.Run("unmarshal JSON to contract", func(t *testing.T) {
		jsonData := `{
			"consumer": {"name": "OrderService"},
			"provider": {"name": "UserService"},
			"interactions": [
				{
					"description": "a request for user 1",
					"providerState": "user 1 exists",
					"request": {
						"method": "GET",
						"path": "/users/1"
					},
					"response": {
						"status": 200,
						"body": {"id": 1, "name": "John Doe"}
					}
				}
			],
			"metadata": {
				"pactSpecification": {"version": "3.0.0"}
			}
		}`

		var c contract.Contract
		err := json.Unmarshal([]byte(jsonData), &c)
		require.NoError(t, err)

		assert.Equal(t, "OrderService", c.Consumer.Name)
		assert.Equal(t, "UserService", c.Provider.Name)
		assert.Len(t, c.Interactions, 1)
		assert.Equal(t, "a request for user 1", c.Interactions[0].Description)
		assert.Equal(t, "user 1 exists", c.Interactions[0].ProviderState)
		assert.Equal(t, "GET", c.Interactions[0].Request.Method)
		assert.Equal(t, "/users/1", c.Interactions[0].Request.Path)
		assert.Equal(t, 200, c.Interactions[0].Response.Status)
		assert.Equal(t, "3.0.0", c.Metadata.PactSpecification.Version)
	})

	t.Run("marshal contract with matching rules", func(t *testing.T) {
		min := 1
		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "test interaction",
					Request:     contract.Request{Method: "GET", Path: "/test"},
					Response: contract.Response{
						Status: 200,
						Body:   map[string]interface{}{"items": []interface{}{}},
						MatchingRules: contract.MatchingRules{
							Body: map[string]contract.MatcherSet{
								"$.items": {
									Matchers: []contract.Matcher{
										{Match: "type", Min: &min},
									},
								},
							},
						},
					},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		data, err := json.Marshal(c)
		require.NoError(t, err)

		var result contract.Contract
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.NotNil(t, result.Interactions[0].Response.MatchingRules.Body["$.items"])
		assert.Equal(t, "type", result.Interactions[0].Response.MatchingRules.Body["$.items"].Matchers[0].Match)
		assert.Equal(t, 1, *result.Interactions[0].Response.MatchingRules.Body["$.items"].Matchers[0].Min)
	})

	t.Run("marshal contract with provider states v3", func(t *testing.T) {
		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "test interaction",
					ProviderStates: []contract.ProviderState{
						{
							Name:   "user exists",
							Params: map[string]interface{}{"userId": float64(1)},
						},
					},
					Request:  contract.Request{Method: "GET", Path: "/users/1"},
					Response: contract.Response{Status: 200},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		data, err := json.Marshal(c)
		require.NoError(t, err)

		var result contract.Contract
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Len(t, result.Interactions[0].ProviderStates, 1)
		assert.Equal(t, "user exists", result.Interactions[0].ProviderStates[0].Name)
		assert.Equal(t, float64(1), result.Interactions[0].ProviderStates[0].Params["userId"])
	})
}

func TestPacticipant_JSONMarshaling(t *testing.T) {
	t.Run("marshal pacticipant", func(t *testing.T) {
		p := contract.Pacticipant{Name: "TestService"}

		data, err := json.Marshal(p)
		require.NoError(t, err)

		assert.JSONEq(t, `{"name":"TestService"}`, string(data))
	})

	t.Run("unmarshal pacticipant", func(t *testing.T) {
		jsonData := `{"name": "TestService"}`

		var p contract.Pacticipant
		err := json.Unmarshal([]byte(jsonData), &p)
		require.NoError(t, err)

		assert.Equal(t, "TestService", p.Name)
	})
}

func TestRequest_JSONMarshaling(t *testing.T) {
	t.Run("marshal request with all fields", func(t *testing.T) {
		r := contract.Request{
			Method: "POST",
			Path:   "/users",
			Query:  map[string][]string{"filter": {"active"}},
			Headers: map[string]interface{}{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{"name": "John"},
		}

		data, err := json.Marshal(r)
		require.NoError(t, err)

		var result contract.Request
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "POST", result.Method)
		assert.Equal(t, "/users", result.Path)
		assert.Equal(t, []string{"active"}, result.Query["filter"])
		assert.Equal(t, "application/json", result.Headers["Content-Type"])
	})
}

func TestResponse_JSONMarshaling(t *testing.T) {
	t.Run("marshal response with all fields", func(t *testing.T) {
		r := contract.Response{
			Status: 201,
			Headers: map[string]interface{}{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"id":   float64(1),
				"name": "John",
			},
		}

		data, err := json.Marshal(r)
		require.NoError(t, err)

		var result contract.Response
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, 201, result.Status)
		assert.Equal(t, "application/json", result.Headers["Content-Type"])
	})
}

func TestMatchingRules_JSONMarshaling(t *testing.T) {
	t.Run("marshal matching rules with body matchers", func(t *testing.T) {
		mr := contract.MatchingRules{
			Body: map[string]contract.MatcherSet{
				"$.id": {
					Matchers: []contract.Matcher{
						{Match: "type"},
					},
				},
				"$.email": {
					Matchers: []contract.Matcher{
						{Match: "regex", Regex: `^[\w.+-]+@[\w.-]+\.\w+$`},
					},
				},
			},
		}

		data, err := json.Marshal(mr)
		require.NoError(t, err)

		var result contract.MatchingRules
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "type", result.Body["$.id"].Matchers[0].Match)
		assert.Equal(t, "regex", result.Body["$.email"].Matchers[0].Match)
		assert.Equal(t, `^[\w.+-]+@[\w.-]+\.\w+$`, result.Body["$.email"].Matchers[0].Regex)
	})

	t.Run("marshal matching rules with combine", func(t *testing.T) {
		mr := contract.MatchingRules{
			Body: map[string]contract.MatcherSet{
				"$.name": {
					Matchers: []contract.Matcher{
						{Match: "type"},
						{Match: "include", Value: "test"},
					},
					Combine: "AND",
				},
			},
		}

		data, err := json.Marshal(mr)
		require.NoError(t, err)

		var result contract.MatchingRules
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Len(t, result.Body["$.name"].Matchers, 2)
		assert.Equal(t, "AND", result.Body["$.name"].Combine)
	})
}

func TestMetadata_JSONMarshaling(t *testing.T) {
	t.Run("marshal metadata with pact specification", func(t *testing.T) {
		m := contract.Metadata{
			PactSpecification: contract.PactSpec{Version: "3.0.0"},
			Client: &contract.Client{
				Name:    "yakusoku",
				Version: "1.0.0",
			},
		}

		data, err := json.Marshal(m)
		require.NoError(t, err)

		var result contract.Metadata
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "3.0.0", result.PactSpecification.Version)
		assert.Equal(t, "yakusoku", result.Client.Name)
		assert.Equal(t, "1.0.0", result.Client.Version)
	})

	t.Run("marshal metadata without client", func(t *testing.T) {
		m := contract.Metadata{
			PactSpecification: contract.PactSpec{Version: "4.0.0"},
		}

		data, err := json.Marshal(m)
		require.NoError(t, err)

		assert.NotContains(t, string(data), "client")
	})
}
