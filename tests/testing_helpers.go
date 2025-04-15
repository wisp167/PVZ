// testing_helpers.go
package tests

import (
	"bytes"
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wisp167/pvz/api"
)

const (
	apiURL = "http://localhost:8080"
)

var (
	validCities  = []string{"Москва", "Санкт-Петербург", "Казань"}
	productTypes = []string{"электроника", "одежда", "обувь"}
)

// GenerateRandomStringSample generates a random string of given length
func GenerateRandomStringSample(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length <= 0 {
		return ""
	}

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var result strings.Builder

	for i := 0; i < length; i++ {
		randomIndex := seededRand.Intn(len(charset))
		result.WriteByte(charset[randomIndex])
	}

	return result.String()
}

// Generate_Username_Password generates username and password pairs
func Generate_Username_Password(i int) (string, string) {
	prefix := GenerateRandomStringSample(6)
	username := prefix + "_" + strconv.Itoa(i)
	password := prefix
	return username, password
}
func makeRequest(t *testing.T, method, url, token string, body []byte) *http.Response {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	assert.NoError(t, err, "Failed to create request")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err, "Request failed")
	return resp
}

// Helper function to authenticate and get a token
func authenticateUser(t *testing.T, role string) string {
	var roleEnum api.PostDummyLoginJSONBodyRole
	switch role {
	case "employee":
		roleEnum = api.PostDummyLoginJSONBodyRoleEmployee
	case "moderator":
		roleEnum = api.PostDummyLoginJSONBodyRoleModerator
	default:
		t.Fatalf("invalid role: %s", role)
	}

	client, err := api.NewClientWithResponses(apiURL)
	assert.NoError(t, err)

	resp, err := client.PostDummyLoginWithResponse(context.Background(), api.PostDummyLoginJSONRequestBody{
		Role: roleEnum,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.NotNil(t, resp.JSON200)

	return string(*resp.JSON200)
}

// Helper to create a PVZ
func createPVZ(t *testing.T, token string, city string) *api.PVZ {
	var cityEnum api.PVZCity
	switch city {
	case "Москва":
		cityEnum = api.PVZCity("Москва")
	case "Санкт-Петербург":
		cityEnum = api.PVZCity("Санкт-Петербург")
	case "Казань":
		cityEnum = api.PVZCity("Казань")
	default:
		t.Fatalf("invalid city: %s", city)
	}

	client, err := api.NewClientWithResponses(apiURL)
	assert.NoError(t, err)

	resp, err := client.PostPvzWithResponse(context.Background(), api.PostPvzJSONRequestBody{
		City: cityEnum,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
	assert.NotNil(t, resp.JSON201)

	return resp.JSON201
}

// Helper to create a reception
func createReception(t *testing.T, token string, pvzID string) *api.Reception {
	client, err := api.NewClientWithResponses(apiURL)
	assert.NoError(t, err)

	pvzUUID, err := uuid.Parse(pvzID)
	assert.NoError(t, err)

	resp, err := client.PostReceptionsWithResponse(context.Background(), api.PostReceptionsJSONRequestBody{
		PvzId: pvzUUID,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
	assert.NotNil(t, resp.JSON201)

	return resp.JSON201
}

// Helper to add a product
func addProduct(t *testing.T, token string, pvzID string, productType string) *api.Product {
	client, err := api.NewClientWithResponses(apiURL)
	assert.NoError(t, err)

	pvzUUID, err := uuid.Parse(pvzID)
	assert.NoError(t, err)

	var typeEnum api.PostProductsJSONBodyType
	switch productType {
	case "электроника":
		typeEnum = api.PostProductsJSONBodyType("электроника")
	case "одежда":
		typeEnum = api.PostProductsJSONBodyType("одежда")
	case "обувь":
		typeEnum = api.PostProductsJSONBodyType("обувь")
	default:
		t.Fatalf("invalid product type: %s", productType)
	}

	resp, err := client.PostProductsWithResponse(context.Background(), api.PostProductsJSONRequestBody{
		PvzId: pvzUUID,
		Type:  typeEnum,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
	assert.NotNil(t, resp.JSON201)

	return resp.JSON201
}

// Helper to close reception
func closeReception(t *testing.T, token string, pvzID string) *api.Reception {
	client, err := api.NewClientWithResponses(apiURL)
	assert.NoError(t, err)

	pvzUUID, err := uuid.Parse(pvzID)
	assert.NoError(t, err)

	resp, err := client.PostPvzPvzIdCloseLastReceptionWithResponse(context.Background(), pvzUUID, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.NotNil(t, resp.JSON200)

	return resp.JSON200
}
