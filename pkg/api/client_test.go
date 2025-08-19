package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMetabaseClient(t *testing.T) {
	client := NewMetabaseClient("https://example.com", "test-token")

	if client.BaseURL != "https://example.com" {
		t.Errorf("NewMetabaseClient() BaseURL = %s, want https://example.com", client.BaseURL)
	}
	if client.APIToken != "test-token" {
		t.Errorf("NewMetabaseClient() APIToken = %s, want test-token", client.APIToken)
	}
	if client.HTTPClient == nil {
		t.Error("NewMetabaseClient() HTTPClient should not be nil")
	}
}

func TestMetabaseClient_TestConnection(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError bool
		errorContains string
	}{
		{
			name:          "successful connection",
			statusCode:    200,
			responseBody:  `{"id": 1, "email": "test@example.com"}`,
			expectedError: false,
		},
		{
			name:          "unauthorized",
			statusCode:    401,
			responseBody:  `{"error": "Invalid API key"}`,
			expectedError: true,
			errorContains: "API token authentication failed",
		},
		{
			name:          "not found",
			statusCode:    404,
			responseBody:  `{"error": "Not found"}`,
			expectedError: true,
			errorContains: "API token authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/user/current" {
					t.Errorf("Expected path /api/user/current, got %s", r.URL.Path)
				}

				apiKey := r.Header.Get("X-API-Key")
				if apiKey != "test-token" {
					t.Errorf("Expected X-API-Key: test-token, got %s", apiKey)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewMetabaseClient(server.URL, "test-token")
			err := client.TestConnection()

			if tt.expectedError {
				if err == nil {
					t.Errorf("TestConnection() expected error, got nil")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("TestConnection() error = %v, want error containing %s", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("TestConnection() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestMetabaseClient_GetDatabases(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedLen   int
		expectedError bool
	}{
		{
			name:       "successful response",
			statusCode: 200,
			responseBody: `{
				"data": [
					{"id": 1, "name": "Sample Database", "engine": "postgres"},
					{"id": 2, "name": "Analytics", "engine": "mysql"}
				]
			}`,
			expectedLen:   2,
			expectedError: false,
		},
		{
			name:          "unauthorized",
			statusCode:    401,
			responseBody:  `{"error": "Unauthorized"}`,
			expectedError: true,
		},
		{
			name:       "empty response",
			statusCode: 200,
			responseBody: `{
				"data": []
			}`,
			expectedLen:   0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/database" {
					t.Errorf("Expected path /api/database, got %s", r.URL.Path)
				}

				apiKey := r.Header.Get("X-API-Key")
				if apiKey != "test-token" {
					t.Errorf("Expected X-API-Key: test-token, got %s", apiKey)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewMetabaseClient(server.URL, "test-token")
			databases, err := client.GetDatabases()

			if tt.expectedError {
				if err == nil {
					t.Errorf("GetDatabases() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetDatabases() unexpected error = %v", err)
				}
				if len(databases) != tt.expectedLen {
					t.Errorf("GetDatabases() returned %d databases, want %d", len(databases), tt.expectedLen)
				}
			}
		})
	}
}

func TestMetabaseClient_GetTables(t *testing.T) {
	tests := []struct {
		name          string
		databaseID    int
		statusCode    int
		responseBody  string
		expectedLen   int
		expectedError bool
	}{
		{
			name:       "successful response",
			databaseID: 1,
			statusCode: 200,
			responseBody: `{
				"tables": [
					{"id": 100, "name": "users", "display_name": "Users", "schema": "public"},
					{"id": 101, "name": "orders", "display_name": "Orders", "schema": "public"}
				]
			}`,
			expectedLen:   2,
			expectedError: false,
		},
		{
			name:          "database not found",
			databaseID:    999,
			statusCode:    404,
			responseBody:  `{"error": "Database not found"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/database/1/metadata"
				if tt.databaseID == 999 {
					expectedPath = "/api/database/999/metadata"
				}
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewMetabaseClient(server.URL, "test-token")
			tables, err := client.GetTables(tt.databaseID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("GetTables() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetTables() unexpected error = %v", err)
				}
				if len(tables) != tt.expectedLen {
					t.Errorf("GetTables() returned %d tables, want %d", len(tables), tt.expectedLen)
				}
			}
		})
	}
}

func TestMetabaseClient_GetTableFields(t *testing.T) {
	tests := []struct {
		name          string
		tableID       int
		statusCode    int
		responseBody  string
		expectedLen   int
		expectedError bool
	}{
		{
			name:       "successful response",
			tableID:    100,
			statusCode: 200,
			responseBody: `{
				"fields": [
					{"id": 200, "name": "id", "display_name": "ID", "base_type": "type/BigInteger"},
					{"id": 201, "name": "email", "display_name": "Email", "base_type": "type/Text"}
				]
			}`,
			expectedLen:   2,
			expectedError: false,
		},
		{
			name:          "table not found",
			tableID:       999,
			statusCode:    404,
			responseBody:  `{"error": "Table not found"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/table/100/query_metadata"
				if tt.tableID == 999 {
					expectedPath = "/api/table/999/query_metadata"
				}
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewMetabaseClient(server.URL, "test-token")
			fields, err := client.GetTableFields(tt.tableID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("GetTableFields() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetTableFields() unexpected error = %v", err)
				}
				if len(fields) != tt.expectedLen {
					t.Errorf("GetTableFields() returned %d fields, want %d", len(fields), tt.expectedLen)
				}
			}
		})
	}
}

func TestMetabaseClient_InvalidBaseURL(t *testing.T) {
	client := NewMetabaseClient("not-a-valid-url", "test-token")

	err := client.TestConnection()
	if err == nil {
		t.Error("TestConnection() with invalid URL should return error")
	}

	_, err = client.GetDatabases()
	if err == nil {
		t.Error("GetDatabases() with invalid URL should return error")
	}

	_, err = client.GetTables(1)
	if err == nil {
		t.Error("GetTables() with invalid URL should return error")
	}

	_, err = client.GetTableFields(1)
	if err == nil {
		t.Error("GetTableFields() with invalid URL should return error")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)) ||
			containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
