// client_test_helpers.go
package tests

import (
	"context"
	"net/http"
	"testing"

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

func NewTestClient(t *testing.T) *api.ClientWithResponses {
	client, err := api.NewClientWithResponses(apiURL)
	assert.NoError(t, err)
	return client
}

func AuthenticateUser(t *testing.T, role string) string {
	client := NewTestClient(t)

	var roleEnum api.PostDummyLoginJSONBodyRole
	switch role {
	case "employee":
		roleEnum = api.PostDummyLoginJSONBodyRoleEmployee
	case "moderator":
		roleEnum = api.PostDummyLoginJSONBodyRoleModerator
	default:
		t.Fatalf("invalid role: %s", role)
	}

	resp, err := client.PostDummyLoginWithResponse(context.Background(), api.PostDummyLoginJSONRequestBody{
		Role: roleEnum,
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())
	assert.NotNil(t, resp.JSON200)

	return string(*resp.JSON200)
}

func CreatePVZ(t *testing.T, token string, city string) *api.PVZ {
	client := NewTestClient(t)

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

	resp, err := client.PostPvzWithResponse(context.Background(), api.PostPvzJSONRequestBody{
		City: cityEnum,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	assert.NotNil(t, resp.JSON201)

	return resp.JSON201
}

func CreateReception(t *testing.T, token string, pvzID string) *api.Reception {
	client := NewTestClient(t)

	resp, err := client.PostReceptionsWithResponse(context.Background(), api.PostReceptionsJSONRequestBody{
		PvzId: pvzID,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	assert.NotNil(t, resp.JSON201)

	return resp.JSON201
}

func AddProduct(t *testing.T, token string, pvzID string, productType string) *api.Product {
	client := NewTestClient(t)

	var typeEnum api.PostProductsJSONBodyType
	switch productType {
	case "электроника":
		typeEnum = api.PostProductsJSONBodyTypeЭлектроника
	case "одежда":
		typeEnum = api.PostProductsJSONBodyTypeОдежда
	case "обувь":
		typeEnum = api.PostProductsJSONBodyTypeОбувь
	default:
		t.Fatalf("invalid product type: %s", productType)
	}

	resp, err := client.PostProductsWithResponse(context.Background(), api.PostProductsJSONRequestBody{
		PvzId: pvzID,
		Type:  typeEnum,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode())
	assert.NotNil(t, resp.JSON201)

	return resp.JSON201
}

func CloseReception(t *testing.T, token string, pvzID string) *api.Reception {
	client := NewTestClient(t)

	resp, err := client.PostPvzPvzIdCloseLastReceptionWithResponse(context.Background(), pvzID, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())
	assert.NotNil(t, resp.JSON200)

	return resp.JSON200
}
