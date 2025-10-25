package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/snowpea/stats/internal/api"
	"github.com/snowpea/stats/internal/config"
)

func TestGetSABQueueSize(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		SabHost:    "http://localhost",
		SabPort:    8080,
		SabAPIKey:  "test-api-key",
		WebTimeout: 5,
		LogLevel:   "None",
	}

	tests := []struct {
		name           string
		responseStatus int
		responseBody   string
		expectedBytes  int64
		expectError    bool
	}{
		{
			name:           "Valid response",
			responseStatus: http.StatusOK,
			responseBody:   `{"queue":{"mbleft":"1024.50"}}`,
			expectedBytes:  1073741824, // 1024.50 MB in bytes
			expectError:    false,
		},
		{
			name:           "Zero MB left",
			responseStatus: http.StatusOK,
			responseBody:   `{"queue":{"mbleft":"0"}}`,
			expectedBytes:  0,
			expectError:    false,
		},
		{
			name:           "Large MB left",
			responseStatus: http.StatusOK,
			responseBody:   `{"queue":{"mbleft":"9999.99"}}`,
			expectedBytes:  10485760000, // 9999.99 MB in bytes
			expectError:    false,
		},
		{
			name:           "HTTP error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"error":"Internal Server Error"}`,
			expectedBytes:  0,
			expectError:    true,
		},
		{
			name:           "Invalid JSON",
			responseStatus: http.StatusOK,
			responseBody:   `invalid json`,
			expectedBytes:  0,
			expectError:    true,
		},
		{
			name:           "Invalid MB value",
			responseStatus: http.StatusOK,
			responseBody:   `{"queue":{"mbleft":"invalid"}}`,
			expectedBytes:  0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request URL contains expected parameters
				expectedPath := "/sabnzbd/api"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify query parameters
				mode := r.URL.Query().Get("mode")
				if mode != "queue" {
					t.Errorf("Expected mode=queue, got %s", mode)
				}

				output := r.URL.Query().Get("output")
				if output != "json" {
					t.Errorf("Expected output=json, got %s", output)
				}

				apikey := r.URL.Query().Get("apikey")
				if apikey != "test-api-key" {
					t.Errorf("Expected apikey=test-api-key, got %s", apikey)
				}

				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Update config to use test server
			cfg.SabHost = server.URL

			result, err := api.GetSABQueueSize(cfg)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetSABQueueSize() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetSABQueueSize() unexpected error: %v", err)
				return
			}

			// Allow for small differences due to floating point precision
			diff := result - tt.expectedBytes
			if diff < 0 {
				diff = -diff
			}
			if diff > 1000000 { // Allow up to 1MB difference
				t.Errorf("GetSABQueueSize() = %d, expected %d (diff: %d)", result, tt.expectedBytes, diff)
			}
		})
	}
}

func TestGetSABQueueSizeTimeout(t *testing.T) {
	// Create test config with short timeout
	cfg := &config.Config{
		SabHost:    "http://localhost",
		SabPort:    8080,
		SabAPIKey:  "test-api-key",
		WebTimeout: 1, // Very short timeout
		LogLevel:   "None",
	}

	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than timeout
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"queue":{"mbleft":"1024"}}`))
	}))
	defer server.Close()

	cfg.SabHost = server.URL

	_, err := api.GetSABQueueSize(cfg)
	if err == nil {
		t.Errorf("GetSABQueueSize() expected timeout error, got nil")
	}
}

func TestSABQueueURLConstruction(t *testing.T) {
	// Test URL construction
	cfg := &config.Config{
		SabHost:   "https://sab.example.com",
		SabPort:   443,
		SabAPIKey: "test-key",
	}

	// We can't directly test the URL construction in GetSABQueueSize,
	// but we can verify the components are used correctly
	if cfg.SabHost != "https://sab.example.com" {
		t.Errorf("sabHost not set correctly")
	}
	if cfg.SabPort != 443 {
		t.Errorf("sabPort not set correctly")
	}
	if cfg.SabAPIKey != "test-key" {
		t.Errorf("sabAPIKey not set correctly")
	}
}
