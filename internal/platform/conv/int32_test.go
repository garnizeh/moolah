//go:build !386 && !arm

package conv

import (
	"math"
	"testing"
)

func TestToInt32(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    int32
		wantErr bool
	}{
		{
			name:    "zero",
			input:   0,
			want:    0,
			wantErr: false,
		},
		{
			name:    "positive small",
			input:   100,
			want:    100,
			wantErr: false,
		},
		{
			name:    "negative small",
			input:   -100,
			want:    -100,
			wantErr: false,
		},
		{
			name:    "max int32",
			input:   math.MaxInt32,
			want:    math.MaxInt32,
			wantErr: false,
		},
		{
			name:    "min int32",
			input:   math.MinInt32,
			want:    math.MinInt32,
			wantErr: false,
		},
		{
			name:    "overflow positive",
			input:   math.MaxInt32 + 1,
			want:    0,
			wantErr: true,
		},
		{
			name:    "overflow negative",
			input:   math.MinInt32 - 1,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt32(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}
