package turbo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedirect(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "root path", url: "/"},
		{name: "nested path", url: "/foo/bar"},
		{name: "absolute url", url: "https://example.com/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			Redirect(w, r, tt.url)

			assert.Equal(t, http.StatusSeeOther, w.Code)
			assert.Equal(t, tt.url, w.Header().Get("Location"))
		})
	}
}
