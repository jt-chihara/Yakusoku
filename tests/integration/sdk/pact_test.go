package sdk_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/sdk/go/yakusoku"
)

func TestPact_NewPact(t *testing.T) {
	t.Run("creates pact with config", func(t *testing.T) {
		tmpDir := t.TempDir()
		pact := yakusoku.NewPact(yakusoku.Config{
			Consumer: "TestConsumer",
			Provider: "TestProvider",
			PactDir:  tmpDir,
		})
		defer pact.Teardown()

		assert.NotNil(t, pact)
		assert.Equal(t, "TestConsumer", pact.Consumer())
		assert.Equal(t, "TestProvider", pact.Provider())
	})
}

func TestPact_InteractionDSL(t *testing.T) {
	t.Run("builds interaction with DSL", func(t *testing.T) {
		tmpDir := t.TempDir()
		pact := yakusoku.NewPact(yakusoku.Config{
			Consumer: "OrderService",
			Provider: "UserService",
			PactDir:  tmpDir,
		})
		defer pact.Teardown()

		pact.
			Given("user 1 exists").
			UponReceiving("a request for user 1").
			WithRequest(yakusoku.Request{
				Method: "GET",
				Path:   "/users/1",
			}).
			WillRespondWith(yakusoku.Response{
				Status: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"id":   1,
					"name": "John Doe",
				},
			})

		assert.True(t, pact.HasInteractions())
	})
}

func TestPact_MockServer(t *testing.T) {
	t.Run("starts mock server for verification", func(t *testing.T) {
		tmpDir := t.TempDir()
		pact := yakusoku.NewPact(yakusoku.Config{
			Consumer: "OrderService",
			Provider: "UserService",
			PactDir:  tmpDir,
		})
		defer pact.Teardown()

		pact.
			Given("user 1 exists").
			UponReceiving("a request for user 1").
			WithRequest(yakusoku.Request{
				Method: "GET",
				Path:   "/users/1",
			}).
			WillRespondWith(yakusoku.Response{
				Status: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"id":   1,
					"name": "John Doe",
				},
			})

		err := pact.Verify(func() error {
			resp, err := http.Get(pact.ServerURL() + "/users/1")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return assert.AnError
			}
			return nil
		})

		require.NoError(t, err)
	})
}

func TestPact_ContractGeneration(t *testing.T) {
	t.Run("generates contract file after verification", func(t *testing.T) {
		tmpDir := t.TempDir()
		pact := yakusoku.NewPact(yakusoku.Config{
			Consumer: "OrderService",
			Provider: "UserService",
			PactDir:  tmpDir,
		})
		defer pact.Teardown()

		pact.
			Given("user 1 exists").
			UponReceiving("a request for user 1").
			WithRequest(yakusoku.Request{
				Method: "GET",
				Path:   "/users/1",
			}).
			WillRespondWith(yakusoku.Response{
				Status: 200,
				Body: map[string]interface{}{
					"id":   1,
					"name": "John Doe",
				},
			})

		err := pact.Verify(func() error {
			resp, err := http.Get(pact.ServerURL() + "/users/1")
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
		})
		require.NoError(t, err)

		// Check contract file was created
		contractPath := filepath.Join(tmpDir, "orderservice-userservice.json")
		_, err = os.Stat(contractPath)
		require.NoError(t, err)

		// Verify contract content
		data, err := os.ReadFile(contractPath)
		require.NoError(t, err)

		var contract map[string]interface{}
		err = json.Unmarshal(data, &contract)
		require.NoError(t, err)

		consumer := contract["consumer"].(map[string]interface{})
		assert.Equal(t, "OrderService", consumer["name"])

		provider := contract["provider"].(map[string]interface{})
		assert.Equal(t, "UserService", provider["name"])

		interactions := contract["interactions"].([]interface{})
		assert.Len(t, interactions, 1)
	})

	t.Run("generates contract with multiple interactions", func(t *testing.T) {
		tmpDir := t.TempDir()
		pact := yakusoku.NewPact(yakusoku.Config{
			Consumer: "OrderService",
			Provider: "UserService",
			PactDir:  tmpDir,
		})
		defer pact.Teardown()

		pact.
			Given("user 1 exists").
			UponReceiving("a request for user 1").
			WithRequest(yakusoku.Request{Method: "GET", Path: "/users/1"}).
			WillRespondWith(yakusoku.Response{Status: 200})

		pact.
			Given("user 2 exists").
			UponReceiving("a request for user 2").
			WithRequest(yakusoku.Request{Method: "GET", Path: "/users/2"}).
			WillRespondWith(yakusoku.Response{Status: 200})

		err := pact.Verify(func() error {
			http.Get(pact.ServerURL() + "/users/1")
			http.Get(pact.ServerURL() + "/users/2")
			return nil
		})
		require.NoError(t, err)

		// Verify contract has 2 interactions
		contractPath := filepath.Join(tmpDir, "orderservice-userservice.json")
		data, err := os.ReadFile(contractPath)
		require.NoError(t, err)

		var contract map[string]interface{}
		json.Unmarshal(data, &contract)

		interactions := contract["interactions"].([]interface{})
		assert.Len(t, interactions, 2)
	})
}

func TestPact_VerifyFailure(t *testing.T) {
	t.Run("returns error when callback fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		pact := yakusoku.NewPact(yakusoku.Config{
			Consumer: "OrderService",
			Provider: "UserService",
			PactDir:  tmpDir,
		})
		defer pact.Teardown()

		pact.
			UponReceiving("a request").
			WithRequest(yakusoku.Request{Method: "GET", Path: "/test"}).
			WillRespondWith(yakusoku.Response{Status: 200})

		err := pact.Verify(func() error {
			return assert.AnError
		})

		require.Error(t, err)
	})
}

func TestPact_Teardown(t *testing.T) {
	t.Run("stops mock server", func(t *testing.T) {
		tmpDir := t.TempDir()
		pact := yakusoku.NewPact(yakusoku.Config{
			Consumer: "OrderService",
			Provider: "UserService",
			PactDir:  tmpDir,
		})

		pact.
			UponReceiving("test").
			WithRequest(yakusoku.Request{Method: "GET", Path: "/test"}).
			WillRespondWith(yakusoku.Response{Status: 200})

		// Start server by calling Verify
		pact.Verify(func() error { return nil })
		serverURL := pact.ServerURL()

		pact.Teardown()

		// Server should no longer be accessible
		_, err := http.Get(serverURL + "/test")
		assert.Error(t, err)
	})
}
