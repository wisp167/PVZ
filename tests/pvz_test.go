package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPVZCreation(t *testing.T) {
	moderatorToken := authenticateUser(t, "moderator")
	employeeToken := authenticateUser(t, "employee")

	tests := []struct {
		name       string
		token      string
		city       string
		wantStatus int
	}{
		{"Moderator creates PVZ in Moscow", moderatorToken, "Москва", http.StatusCreated},
		{"Moderator creates PVZ in SPb", moderatorToken, "Санкт-Петербург", http.StatusCreated},
		{"Moderator creates PVZ in Kazan", moderatorToken, "Казань", http.StatusCreated},
		{"Employee tries to create PVZ", employeeToken, "Москва", http.StatusForbidden},
		{"Invalid city", moderatorToken, "Новосибирск", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvz := map[string]string{"city": tt.city}
			body, _ := json.Marshal(pvz)

			resp := makeRequest(t, "POST", apiURL+"/pvz", tt.token, body)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestGetPVZs(t *testing.T) {
	moderatorToken := authenticateUser(t, "moderator")

	// Create some PVZs first
	for _, city := range validCities {
		createPVZ(t, moderatorToken, city)
	}

	// Test getting all PVZs
	resp := makeRequest(t, "GET", apiURL+"/pvz", moderatorToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with date filters
	now := time.Now()
	startDate := now.Add(-24 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	url := fmt.Sprintf("%s/pvz?startDate=%s&endDate=%s", apiURL, startDate, endDate)
	resp = makeRequest(t, "GET", url, moderatorToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
