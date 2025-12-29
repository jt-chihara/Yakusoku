package contract_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

func TestValidator_Validate(t *testing.T) {
	validContract := func() contract.Contract {
		return contract.Contract{
			Consumer: contract.Pacticipant{Name: "OrderService"},
			Provider: contract.Pacticipant{Name: "UserService"},
			Interactions: []contract.Interaction{
				{
					Description: "a request for user 1",
					Request: contract.Request{
						Method: "GET",
						Path:   "/users/1",
					},
					Response: contract.Response{
						Status: 200,
					},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}
	}

	t.Run("valid contract passes validation", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()

		err := v.Validate(c)
		require.NoError(t, err)
	})

	t.Run("missing consumer name fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Consumer.Name = ""

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "consumer name")
	})

	t.Run("missing provider name fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Provider.Name = ""

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider name")
	})

	t.Run("empty interactions fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions = []contract.Interaction{}

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "interaction")
	})

	t.Run("nil interactions fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions = nil

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "interaction")
	})

	t.Run("missing interaction description fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Description = ""

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "description")
	})

	t.Run("invalid HTTP method fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Request.Method = "INVALID"

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "method")
	})

	t.Run("missing request path fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Request.Path = ""

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path")
	})

	t.Run("path not starting with slash fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Request.Path = "users/1"

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path")
	})

	t.Run("invalid status code fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Response.Status = 0

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status")
	})

	t.Run("status code below 100 fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Response.Status = 99

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status")
	})

	t.Run("status code above 599 fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Response.Status = 600

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status")
	})

	t.Run("consumer name too long fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Consumer.Name = string(make([]byte, 256))

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "consumer name")
	})

	t.Run("all valid HTTP methods pass", func(t *testing.T) {
		validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		v := contract.NewValidator()

		for _, method := range validMethods {
			c := validContract()
			c.Interactions[0].Request.Method = method

			err := v.Validate(c)
			assert.NoError(t, err, "method %s should be valid", method)
		}
	})

	t.Run("lowercase HTTP method is normalized", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions[0].Request.Method = "get"

		err := v.Validate(c)
		require.NoError(t, err)
	})

	t.Run("multiple interactions all validated", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions = append(c.Interactions, contract.Interaction{
			Description: "another request",
			Request: contract.Request{
				Method: "POST",
				Path:   "/users",
			},
			Response: contract.Response{
				Status: 201,
			},
		})

		err := v.Validate(c)
		require.NoError(t, err)
	})

	t.Run("second interaction invalid fails", func(t *testing.T) {
		v := contract.NewValidator()
		c := validContract()
		c.Interactions = append(c.Interactions, contract.Interaction{
			Description: "",
			Request: contract.Request{
				Method: "POST",
				Path:   "/users",
			},
			Response: contract.Response{
				Status: 201,
			},
		})

		err := v.Validate(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "description")
	})
}

func TestValidator_ValidateRequest(t *testing.T) {
	t.Run("valid request passes", func(t *testing.T) {
		v := contract.NewValidator()
		r := contract.Request{
			Method: "GET",
			Path:   "/users/1",
			Headers: map[string]interface{}{
				"Accept": "application/json",
			},
		}

		err := v.ValidateRequest(r)
		require.NoError(t, err)
	})

	t.Run("request with body passes", func(t *testing.T) {
		v := contract.NewValidator()
		r := contract.Request{
			Method: "POST",
			Path:   "/users",
			Body:   map[string]interface{}{"name": "John"},
		}

		err := v.ValidateRequest(r)
		require.NoError(t, err)
	})

	t.Run("request with query params passes", func(t *testing.T) {
		v := contract.NewValidator()
		r := contract.Request{
			Method: "GET",
			Path:   "/users",
			Query:  map[string][]string{"status": {"active", "pending"}},
		}

		err := v.ValidateRequest(r)
		require.NoError(t, err)
	})
}

func TestValidator_ValidateResponse(t *testing.T) {
	t.Run("valid response passes", func(t *testing.T) {
		v := contract.NewValidator()
		r := contract.Response{
			Status: 200,
			Headers: map[string]interface{}{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{"id": 1},
		}

		err := v.ValidateResponse(r)
		require.NoError(t, err)
	})

	t.Run("response with matching rules passes", func(t *testing.T) {
		v := contract.NewValidator()
		r := contract.Response{
			Status: 200,
			Body:   map[string]interface{}{"id": 1},
			MatchingRules: contract.MatchingRules{
				Body: map[string]contract.MatcherSet{
					"$.id": {
						Matchers: []contract.Matcher{
							{Match: "type"},
						},
					},
				},
			},
		}

		err := v.ValidateResponse(r)
		require.NoError(t, err)
	})

	t.Run("all valid status codes pass", func(t *testing.T) {
		v := contract.NewValidator()
		validStatuses := []int{100, 200, 201, 204, 301, 400, 401, 404, 500, 502, 599}

		for _, status := range validStatuses {
			r := contract.Response{Status: status}
			err := v.ValidateResponse(r)
			assert.NoError(t, err, "status %d should be valid", status)
		}
	})
}
