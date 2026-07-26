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

func TestScriptImport(t *testing.T) {
	tests := []struct {
		name     string
		contains []string
	}{
		{
			name: "renders a module script tag pinned to turboVersion",
			contains: []string{
				`<script type="module"`,
				`@hotwired/turbo@` + turboVersion,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(ScriptImport())
			for _, sub := range tt.contains {
				assert.Contains(t, got, sub)
			}
		})
	}
}

func TestTemplateFuncMap(t *testing.T) {
	// Verify every expected helper is registered under its template
	// funcname; per-helper tests already cover the underlying Go
	// functions, so this only guards against typos or missed
	// registrations in TemplateFuncMap.
	tests := []struct {
		name     string
		wantKeys []string
	}{
		{
			name: "every helper is registered",
			wantKeys: []string{
				"turboScriptImport",
				"turboAttrConfirm",
				"turboAttrSubmitsWith",
				"turboAttrActionAdvance",
				"turboAttrActionReplace",
				"turboAttrMethodDelete",
				"turboAttrMethodPatch",
				"turboAttrMethodPut",
				"turboMetaVisitControlReload",
				"turboMetaCacheControlNoPreview",
				"turboMetaCacheControlNoCache",
				"turboMetaViewTransition",
				"turboMetaRefreshMethodMorph",
				"turboMetaRefreshScrollPreserve",
				"turboMetaDisablePrefetch",
				"turboMetaRoot",
				"turboAttrTrackReload",
				"turboAttrTrackDynamic",
				"turboAttrDisableTurbo",
				"turboAttrEnableTurbo",
				"turboAttrPreload",
				"turboAttrDisablePrefetch",
				"turboAttrPermanent",
				"turboAttrTemporary",
				"turboAttrDisableEval",
				"turboFrame",
				"turboFrameEnd",
				"turboAttrSrc",
				"turboAttrLoadingLazy",
				"turboAttrLoadingEager",
				"turboAttrDisabled",
				"turboAttrTarget",
				"turboAttrFrame",
				"turboAttrRecurse",
				"turboAttrAutoscroll",
				"turboAttrAutoscrollBlockStart",
				"turboAttrAutoscrollBlockCenter",
				"turboAttrAutoscrollBlockEnd",
				"turboAttrAutoscrollBlockNearest",
				"turboAttrAutoscrollBehaviorAuto",
				"turboAttrAutoscrollBehaviorSmooth",
				"turboAttrRefreshMorph",
				"turboStreamAppend",
				"turboStreamPrepend",
				"turboStreamReplace",
				"turboStreamUpdate",
				"turboStreamRemove",
				"turboStreamBefore",
				"turboStreamAfter",
				"turboStreamRefresh",
				"turboStreamEnd",
				"turboAttrTargets",
				"turboAttrRequestID",
				"turboAttrMethodMorph",
				"turboStreamSourceSSE",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := TemplateFuncMap()
			assert.NotEmpty(t, fm)
			for _, key := range tt.wantKeys {
				_, ok := fm[key]
				assert.Truef(t, ok, "funcmap entry %q missing", key)
			}
		})
	}
}

func TestTurboAttrHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  Attrs
		want Attrs
	}{
		{
			name: "AttrConfirm",
			got:  AttrConfirm("Really?"),
			want: Attrs{{Key: "data-turbo-confirm", Value: "Really?"}},
		},
		{
			name: "AttrSubmitsWith",
			got:  AttrSubmitsWith("Saving..."),
			want: Attrs{{Key: "data-turbo-submits-with", Value: "Saving..."}},
		},
		{
			name: "AttrActionAdvance",
			got:  AttrActionAdvance(),
			want: Attrs{{Key: "data-turbo-action", Value: "advance"}},
		},
		{
			name: "AttrActionReplace",
			got:  AttrActionReplace(),
			want: Attrs{{Key: "data-turbo-action", Value: "replace"}},
		},
		{
			name: "AttrMethodDelete",
			got:  AttrMethodDelete(),
			want: Attrs{{Key: "data-turbo-method", Value: "delete"}},
		},
		{
			name: "AttrMethodPatch",
			got:  AttrMethodPatch(),
			want: Attrs{{Key: "data-turbo-method", Value: "patch"}},
		},
		{
			name: "AttrMethodPut",
			got:  AttrMethodPut(),
			want: Attrs{{Key: "data-turbo-method", Value: "put"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestTurboMetaTagHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  Tag
		want Tag
	}{
		{
			name: "MetaVisitControlReload",
			got:  MetaVisitControlReload(),
			want: `<meta name="turbo-visit-control" content="reload">`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
