package view

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartial_Render(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		ctx     context.Context
		partial string
		data    map[string]any
		want    string
	}{
		{
			name:    "shared partial rendered without layout",
			dir:     "testdata/valid/ok",
			ctx:     context.Background(),
			partial: "shared",
			data:    nil,
			want:    "SHARED",
		},
		{
			name:    "partial receives data",
			dir:     "testdata/valid/ok",
			ctx:     context.Background(),
			partial: "greet",
			data:    map[string]any{"Name": "Bob"},
			want:    "Hello Bob",
		},
		{
			name:    "partial in a nested directory is reachable",
			dir:     "testdata/valid/ok",
			ctx:     context.Background(),
			partial: "local",
			data:    nil,
			want:    "LOCAL",
		},
		{
			name:    ".Ctx is exposed to the template",
			dir:     "testdata/valid/ctx",
			ctx:     context.Background(),
			partial: "ctx_probe",
			data:    nil,
			want:    "has-ctx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(testdataFS, tt.dir, WithFuncs(testFuncs))
			require.NoError(t, err)
			p := r.Partial(tt.partial, tt.data)
			var buf bytes.Buffer
			err = p.Render(tt.ctx, &buf)
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
			assert.NotContains(t, tt.data, "Ctx", "Render must not mutate caller's data map")
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
