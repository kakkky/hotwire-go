package view

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartial_Render(t *testing.T) {
	r, err := New(testdataFS, "testdata/valid/ok", WithFuncs(testFuncs))
	require.NoError(t, err)

	tests := []struct {
		name    string
		partial string
		data    any
		want    string
	}{
		{
			name:    "shared partial rendered without layout",
			partial: "shared",
			data:    nil,
			want:    "SHARED",
		},
		{
			name:    "partial receives data",
			partial: "greet",
			data:    map[string]string{"Name": "Bob"},
			want:    "Hello Bob",
		},
		{
			name:    "partial in a nested directory is reachable",
			partial: "local",
			data:    nil,
			want:    "LOCAL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := r.Partial(tt.partial, tt.data)
			var buf bytes.Buffer
			err := p.Render(context.Background(), &buf)
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestPartial_Render_Error(t *testing.T) {
	r, err := New(testdataFS, "testdata/valid/ok", WithFuncs(testFuncs))
	require.NoError(t, err)

	tests := []struct {
		name       string
		partial    string
		wantErrMsg string
	}{
		{
			name:       "unknown partial",
			partial:    "does_not_exist",
			wantErrMsg: `partial "does_not_exist" not found`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := r.Partial(tt.partial, nil)
			var buf bytes.Buffer
			err := p.Render(context.Background(), &buf)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
			assert.Empty(t, buf.String())
		})
	}
}
