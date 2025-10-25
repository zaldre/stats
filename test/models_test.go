package test

import (
	"encoding/json"
	"testing"

	"github.com/snowpea/stats/internal/models"
)

func TestSABQueue(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected string
	}{
		{
			name:     "Valid SAB queue with MB left",
			jsonData: `{"queue":{"mbleft":"1024.50"}}`,
			expected: "1024.50",
		},
		{
			name:     "Empty queue",
			jsonData: `{"queue":{"mbleft":"0"}}`,
			expected: "0",
		},
		{
			name:     "Large queue",
			jsonData: `{"queue":{"mbleft":"999999.99"}}`,
			expected: "999999.99",
		},
		{
			name:     "Decimal queue",
			jsonData: `{"queue":{"mbleft":"123.45"}}`,
			expected: "123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sabQueue models.SABQueue
			err := json.Unmarshal([]byte(tt.jsonData), &sabQueue)
			if err != nil {
				t.Errorf("Failed to unmarshal JSON: %v", err)
				return
			}

			if sabQueue.Queue.MBLeft != tt.expected {
				t.Errorf("SABQueue.Queue.MBLeft = %s, expected %s", sabQueue.Queue.MBLeft, tt.expected)
			}
		})
	}
}

func TestSABQueueJSONUnmarshal(t *testing.T) {
	// Test with complete JSON structure
	jsonData := `{
		"queue": {
			"mbleft": "2048.75"
		}
	}`

	var sabQueue models.SABQueue
	err := json.Unmarshal([]byte(jsonData), &sabQueue)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if sabQueue.Queue.MBLeft != "2048.75" {
		t.Errorf("Expected MBLeft to be '2048.75', got '%s'", sabQueue.Queue.MBLeft)
	}
}

func TestSABQueueJSONMarshal(t *testing.T) {
	sabQueue := models.SABQueue{}
	sabQueue.Queue.MBLeft = "512.25"

	jsonData, err := json.Marshal(sabQueue)
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
	}

	expected := `{"queue":{"mbleft":"512.25"}}`
	if string(jsonData) != expected {
		t.Errorf("Expected JSON: %s, got: %s", expected, string(jsonData))
	}
}
