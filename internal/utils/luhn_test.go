package utils

import (
	"strings"
	"testing"
)

func TestLuhnCheck(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     bool
		wantErr  bool
		errMatch string
	}{
		{
			name:    "valid1",
			input:   "4242424242424242",
			want:    true,
			wantErr: false,
		},
		{
			name:    "valid2",
			input:   "5555555555554444",
			want:    true,
			wantErr: false,
		},
		{
			name:    "valid3",
			input:   "378282246310005",
			want:    true,
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "4242424242424241",
			want:    false,
			wantErr: false,
		},
		{
			name:     "invalid short",
			input:    "12345",
			want:     false,
			wantErr:  true,
			errMatch: "invalid len format",
		},
		{
			name:     "empty string",
			input:    "",
			want:     false,
			wantErr:  true,
			errMatch: "empty value",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			want:     false,
			wantErr:  true,
			errMatch: "empty value",
		},
		{
			name:     "non-numeric",
			input:    "4242a24242424242",
			want:     false,
			wantErr:  true,
			errMatch: "invalid integer format",
		},
		{
			name:     "too long",
			input:    "12345678901234567890",
			want:     false,
			wantErr:  true,
			errMatch: "bad luhn: invalid integer format",
		},
		{
			name:     "with spaces",
			input:    " 4242 4242 4242 4242 ",
			want:     false,
			wantErr:  true,
			errMatch: "bad luhn: invalid integer format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LuhnCheck(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("LuhnCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMatch != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("LuhnCheck() error = %v, want error containing %q", err, tt.errMatch)
				}
			}

			if got != tt.want {
				t.Errorf("LuhnCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
