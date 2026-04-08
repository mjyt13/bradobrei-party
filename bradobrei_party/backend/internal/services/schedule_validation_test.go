package services

import (
	"strings"
	"testing"
)

func TestNormalizeScheduleJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "valid schedule",
			input: `{"mon":"10:00-20:00","tue":"09:30-18:45"}`,
		},
		{
			name:    "letters instead of time",
			input:   `{"mon":"aa:bb-cc:dd"}`,
			wantErr: true,
		},
		{
			name:    "unknown day",
			input:   `{"monday":"10:00-20:00"}`,
			wantErr: true,
		},
		{
			name:    "start after end",
			input:   `{"mon":"20:00-10:00"}`,
			wantErr: true,
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantErr: true,
		},
		{
			name:    "not an object",
			input:   `["10:00-20:00"]`,
			wantErr: true,
		},
		{
			name:    "broken json",
			input:   `{"mon":"10:00-20:00"`,
			wantErr: true,
		},
		{
			name:  "blank string",
			input: `   `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeScheduleJSON(tt.input, "schedule")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(tt.input) == "" {
				if got != nil {
					t.Fatalf("expected nil for blank schedule, got %q", *got)
				}
				return
			}
			if got == nil || !strings.Contains(*got, "10:00-20:00") {
				t.Fatalf("unexpected normalized schedule: %#v", got)
			}
		})
	}
}
