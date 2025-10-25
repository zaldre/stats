package test

import (
	"testing"

	"github.com/snowpea/stats/internal/utils"
)

func TestCalcSize(t *testing.T) {
	tests := []struct {
		name      string
		byteCount int64
		expected  string
		expectErr bool
	}{
		{
			name:      "Zero bytes",
			byteCount: 0,
			expected:  "0 Bytes",
			expectErr: false,
		},
		{
			name:      "Negative bytes",
			byteCount: -1,
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Single byte",
			byteCount: 1,
			expected:  "1 Bytes",
			expectErr: false,
		},
		{
			name:      "1023 bytes",
			byteCount: 1023,
			expected:  "1023 Bytes",
			expectErr: false,
		},
		{
			name:      "1 KB",
			byteCount: 1024,
			expected:  "1.00 KB",
			expectErr: false,
		},
		{
			name:      "1.5 KB",
			byteCount: 1536,
			expected:  "1.50 KB",
			expectErr: false,
		},
		{
			name:      "1 MB",
			byteCount: 1024 * 1024,
			expected:  "1.00 MB",
			expectErr: false,
		},
		{
			name:      "1.5 MB",
			byteCount: 1024 * 1024 * 1.5,
			expected:  "1.50 MB",
			expectErr: false,
		},
		{
			name:      "1 GB",
			byteCount: 1024 * 1024 * 1024,
			expected:  "1.00 GB",
			expectErr: false,
		},
		{
			name:      "1 TB",
			byteCount: 1024 * 1024 * 1024 * 1024,
			expected:  "1.00 TB",
			expectErr: false,
		},
		{
			name:      "Large number (2.5 TB)",
			byteCount: 1024 * 1024 * 1024 * 1024 * 2.5,
			expected:  "2.50 TB",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.CalcSize(tt.byteCount)

			if tt.expectErr {
				if err == nil {
					t.Errorf("CalcSize(%d) expected error, got nil", tt.byteCount)
				}
				return
			}

			if err != nil {
				t.Errorf("CalcSize(%d) unexpected error: %v", tt.byteCount, err)
				return
			}

			if result != tt.expected {
				t.Errorf("CalcSize(%d) = %s, expected %s", tt.byteCount, result, tt.expected)
			}
		})
	}
}
