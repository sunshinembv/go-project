package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatus(t *testing.T) {
	type test struct {
		name    string
		value   string
		want    Status
		wantErr bool
	}

	tests := []test{
		{
			name:  "new",
			value: "new",
			want:  NewStatus,
		},
		{
			name:  "in progress",
			value: "inProgress",
			want:  InProgressStatus,
		},
		{
			name:  "completed",
			value: "completed",
			want:  CompletedStatus,
		},
		{
			name:    "unknown",
			value:   "cancelled",
			want:    -1,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, err := ParseStatus(tc.value)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.value)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, status)
			assert.Equal(t, tc.value, status.String())
		})
	}
}

func TestScan(t *testing.T) {
	type test struct {
		name    string
		value   any
		want    Status
		wantErr bool
	}

	tests := []test{
		{
			name:  "string",
			value: "new",
			want:  NewStatus,
		},
		{
			name:  "bytes",
			value: []byte("completed"),
			want:  CompletedStatus,
		},
		{
			name:    "invalid string",
			value:   "unknown",
			wantErr: true,
		},
		{
			name:    "invalid bytes",
			value:   []byte("unknown"),
			wantErr: true,
		},
		{
			name:    "unsupported type",
			value:   1,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var status Status
			err := status.Scan(tc.value)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, status)
		})
	}
}

func TestStatusJSON(t *testing.T) {
	type test struct {
		name    string
		data    string
		want    Status
		wantErr bool
	}

	tests := []test{
		{
			name: "valid",
			data: `"inProgress"`,
			want: InProgressStatus,
		},
		{
			name:    "unknown",
			data:    `"unknown"`,
			wantErr: true,
		},
		{
			name:    "invalid json type",
			data:    `1`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var status Status
			err := json.Unmarshal([]byte(tc.data), &status)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, status)

			data, err := json.Marshal(status)
			require.NoError(t, err)
			assert.JSONEq(t, tc.data, string(data))
		})
	}
}
