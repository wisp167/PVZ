package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReceptionWorkflow(t *testing.T) {
	moderatorToken := authenticateUser(t, "moderator")
	employeeToken := authenticateUser(t, "employee")

	// Test creating a reception
	t.Run("Create reception", func(t *testing.T) {
		pvz := createPVZ(t, moderatorToken, "Москва")
		reception := createReception(t, employeeToken, pvz.Id.String())
		assert.Equal(t, "in_progress", string(reception.Status))
	})

	// Test cannot create another reception while one is open
	t.Run("Only one open reception", func(t *testing.T) {
		pvz := createPVZ(t, moderatorToken, "Москва")
		createReception(t, employeeToken, pvz.Id.String())
		reqBody := map[string]string{"pvzId": pvz.Id.String()}
		body, _ := json.Marshal(reqBody)

		resp := makeRequest(t, "POST", apiURL+"/receptions", employeeToken, body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test adding products
	t.Run("Add products to reception", func(t *testing.T) {
		pvz := createPVZ(t, moderatorToken, "Москва")
		createReception(t, employeeToken, pvz.Id.String())
		for _, productType := range productTypes {
			product := addProduct(t, employeeToken, pvz.Id.String(), productType)
			assert.Equal(t, productType, string(product.Type))
		}
	})

	// Test closing reception
	t.Run("Close reception", func(t *testing.T) {
		pvz := createPVZ(t, moderatorToken, "Москва")
		createReception(t, employeeToken, pvz.Id.String())
		closedReception := closeReception(t, employeeToken, pvz.Id.String())
		assert.Equal(t, "close", string(closedReception.Status))
	})

	// Test can create new reception after closing
	t.Run("Create new reception after closing", func(t *testing.T) {
		pvz := createPVZ(t, moderatorToken, "Москва")
		createReception(t, employeeToken, pvz.Id.String())
		closeReception(t, employeeToken, pvz.Id.String())
		reception := createReception(t, employeeToken, pvz.Id.String())
		assert.Equal(t, "in_progress", string(reception.Status))
	})
}
