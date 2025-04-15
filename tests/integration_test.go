package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wisp167/pvz/api"
)

func TestIntegrationWorkflow(t *testing.T) {
	moderatorToken := authenticateUser(t, "moderator")

	pvz := createPVZ(t, moderatorToken, "Казань")
	assert.Equal(t, "Казань", string(pvz.City))

	employeeToken := authenticateUser(t, "employee")

	reception := createReception(t, employeeToken, pvz.Id.String())
	assert.Equal(t, "in_progress", string(reception.Status))

	for i := 0; i < 50; i++ {
		productType := productTypes[i%len(productTypes)]
		product := addProduct(t, employeeToken, pvz.Id.String(), productType)
		assert.Equal(t, productType, string(product.Type))
	}

	// 6. Close the reception
	closedReception := closeReception(t, employeeToken, pvz.Id.String())
	assert.Equal(t, "close", string(closedReception.Status))

	// 7. Verify data through GET /pvz
	resp := makeRequest(t, "GET", apiURL+"/pvz", moderatorToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pvzData []struct {
		Pvz        *api.PVZ `json:"pvz"`
		Receptions *[]struct {
			Reception *api.Reception `json:"reception"`
			Products  *[]api.Product `json:"products"`
		} `json:"receptions"`
	}

	err := json.NewDecoder(resp.Body).Decode(&pvzData)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(pvzData), 1)

	// Find our PVZ in the response
	var foundPVZ *api.PVZ
	var foundReceptions []api.Reception
	var foundProducts []api.Product

	for _, item := range pvzData {
		if item.Pvz.Id.String() == pvz.Id.String() {
			foundPVZ = item.Pvz
			for _, rec := range *item.Receptions {
				if rec.Reception.Id.String() == reception.Id.String() {
					foundReceptions = append(foundReceptions, *rec.Reception)
					foundProducts = append(foundProducts, *rec.Products...)
				}
			}
			break
		}
	}

	assert.NotNil(t, foundPVZ)
	assert.Equal(t, 1, len(foundReceptions))
	assert.Equal(t, 50, len(foundProducts))
}
