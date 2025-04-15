package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/wisp167/pvz/api"
)

func TestDummyLogin(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{"Employee login", "employee", http.StatusOK},
		{"Moderator login", "moderator", http.StatusOK},
		{"Invalid role", "invalid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var roleEnum api.PostDummyLoginJSONBodyRole
			switch tt.role {
			case "employee":
				roleEnum = api.PostDummyLoginJSONBodyRoleEmployee
			case "moderator":
				roleEnum = api.PostDummyLoginJSONBodyRoleModerator
			default:
				roleEnum = "invalid"
			}

			client, err := api.NewClientWithResponses(apiURL)
			assert.NoError(t, err)

			resp, err := client.PostDummyLoginWithResponse(context.Background(), api.PostDummyLoginJSONRequestBody{
				Role: roleEnum,
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())

			if tt.wantStatus == http.StatusOK {
				assert.NotNil(t, resp.JSON200)
				assert.NotEmpty(t, *resp.JSON200)
			} else if tt.wantStatus == http.StatusBadRequest {
				assert.NotNil(t, resp.JSON400)
			}
		})
	}
}
func TestRegisterAndLogin(t *testing.T) {
	client, err := api.NewClientWithResponses(apiURL)
	assert.NoError(t, err)

	email := types.Email(GenerateRandomStringSample(4) + "@gmail.com")
	password := GenerateRandomStringSample(10)
	role := "employee"

	registerResp, err := client.PostRegisterWithResponse(context.Background(), api.PostRegisterJSONRequestBody{
		Email:    email,
		Password: password,
		Role:     api.PostRegisterJSONBodyRole(role),
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, registerResp.StatusCode())

	loginResp, err := client.PostLoginWithResponse(context.Background(), api.PostLoginJSONRequestBody{
		Email:    email,
		Password: password,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode())

	invalidLoginResp, err := client.PostLoginWithResponse(context.Background(), api.PostLoginJSONRequestBody{
		Email:    email,
		Password: "wrong",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, invalidLoginResp.StatusCode())
}
