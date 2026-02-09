package parser

import "testing"

func TestParseCPUToMillicores(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"100m", 100, false},
		{"500m", 500, false},
		{"1000m", 1000, false},
		{"0.5", 500, false},
		{"1", 1000, false},
		{"2.5", 2500, false},
		{"0.1", 100, false},
		{"", 0, false},
		{"  250m  ", 250, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := ParseCPUToMillicores(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseCPUToMillicores(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if result != tt.expected {
			t.Errorf("ParseCPUToMillicores(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestParseMemoryToMiB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"128Mi", 128, false},
		{"512Mi", 512, false},
		{"1Gi", 1024, false},
		{"2Gi", 2048, false},
		{"1024Ki", 1, false},
		{"2048Ki", 2, false},
		{"1Ti", 1048576, false},
		{"", 0, false},
		{"  256Mi  ", 256, false},
		{"100M", 95, false}, // 100MB ≈ 95.37 MiB
		{"1G", 953, false},  // 1GB ≈ 953.67 MiB
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := ParseMemoryToMiB(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseMemoryToMiB(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if result != tt.expected {
			t.Errorf("ParseMemoryToMiB(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
