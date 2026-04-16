package utils

import "testing"

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid number - simple",
			number: "0",
			want:   true,
		},
		{
			name:   "valid number - credit card",
			number: "4532015112830366",
			want:   true,
		},
		{
			name:   "valid number - another example",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "invalid number",
			number: "4532015112830367",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "contains non-digit character",
			number: "453201511283036a",
			want:   false,
		},
		{
			name:   "single digit valid",
			number: "0",
			want:   true,
		},
		{
			name:   "single digit invalid",
			number: "1",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateLuhn(tt.number); got != tt.want {
				t.Errorf("ValidateLuhn(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}
