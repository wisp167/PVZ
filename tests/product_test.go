package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductOperations(t *testing.T) {
	moderatorToken := authenticateUser(t, "moderator")
	employeeToken := authenticateUser(t, "employee")

	// Create a PVZ and reception first
	pvz := createPVZ(t, moderatorToken, "Санкт-Петербург")
	createReception(t, employeeToken, pvz.Id.String())

	// Test adding products
	var products []string
	for _, productType := range productTypes {
		t.Run("Add "+productType, func(t *testing.T) {
			product := addProduct(t, employeeToken, pvz.Id.String(), productType)
			assert.Equal(t, productType, string(product.Type))
			products = append(products, product.Id.String())
		})
	}

	// Test delete last product (LIFO)
	t.Run("Delete last product", func(t *testing.T) {
		resp := makeRequest(t, "POST",
			fmt.Sprintf("%s/pvz/%s/delete_last_product", apiURL, pvz.Id.String()),
			employeeToken, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test cannot delete when no products
	t.Run("Delete from empty reception", func(t *testing.T) {
		// Create new empty reception
		closeReception(t, employeeToken, pvz.Id.String())
		createReception(t, employeeToken, pvz.Id.String())

		resp := makeRequest(t, "POST",
			fmt.Sprintf("%s/pvz/%s/delete_last_product", apiURL, pvz.Id.String()),
			employeeToken, nil)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test cannot add to closed reception
	t.Run("Add to closed reception", func(t *testing.T) {
		closeReception(t, employeeToken, pvz.Id.String())

		reqBody := map[string]interface{}{
			"pvzId": pvz.Id.String(),
			"type":  productTypes[0],
		}
		body, _ := json.Marshal(reqBody)

		resp := makeRequest(t, "POST", apiURL+"/products", employeeToken, body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
