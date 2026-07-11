package turbo

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
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
	// All Turbo helpers are exposed through TemplateFuncMap; wantKeys is
	// the full list of registered names, and templateSrc invokes every
	// entry so the closure bodies (not just the map lookups) execute.
	wantKeys := []string{
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
	}

	templateSrc := strings.Join([]string{
		`{{ turboScriptImport }}`,
		`{{ turboAttrConfirm "ok?" }}`,
		`{{ turboAttrSubmitsWith "Saving..." }}`,
		`{{ turboAttrActionAdvance }}`,
		`{{ turboAttrActionReplace }}`,
		`{{ turboAttrMethodDelete }}`,
		`{{ turboAttrMethodPatch }}`,
		`{{ turboAttrMethodPut }}`,
		`{{ turboMetaVisitControlReload }}`,
		`{{ turboMetaCacheControlNoPreview }}`,
		`{{ turboMetaCacheControlNoCache }}`,
		`{{ turboMetaViewTransition }}`,
		`{{ turboMetaRefreshMethodMorph }}`,
		`{{ turboMetaRefreshScrollPreserve }}`,
		`{{ turboMetaDisablePrefetch }}`,
		`{{ turboMetaRoot "/app" }}`,
		`{{ turboAttrTrackReload }}`,
		`{{ turboAttrTrackDynamic }}`,
		`{{ turboAttrDisableTurbo }}`,
		`{{ turboAttrEnableTurbo }}`,
		`{{ turboAttrPreload }}`,
		`{{ turboAttrDisablePrefetch }}`,
		`{{ turboAttrPermanent }}`,
		`{{ turboAttrTemporary }}`,
		`{{ turboAttrDisableEval }}`,
		`{{ turboFrame "x" }}inside{{ turboFrameEnd }}`,
		`{{ turboAttrSrc "/x" }}`,
		`{{ turboAttrLoadingLazy }}`,
		`{{ turboAttrLoadingEager }}`,
		`{{ turboAttrDisabled }}`,
		`{{ turboAttrTarget "_top" }}`,
		`{{ turboAttrFrame "_top" }}`,
		`{{ turboAttrRecurse "nested" }}`,
		`{{ turboAttrAutoscroll }}`,
		`{{ turboAttrAutoscrollBlockStart }}`,
		`{{ turboAttrAutoscrollBlockCenter }}`,
		`{{ turboAttrAutoscrollBlockEnd }}`,
		`{{ turboAttrAutoscrollBlockNearest }}`,
		`{{ turboAttrAutoscrollBehaviorAuto }}`,
		`{{ turboAttrAutoscrollBehaviorSmooth }}`,
		`{{ turboAttrRefreshMorph }}`,
		`{{ turboStreamAppend "m" }}{{ turboStreamEnd }}`,
		`{{ turboStreamPrepend "m" }}{{ turboStreamEnd }}`,
		`{{ turboStreamReplace "m" }}{{ turboStreamEnd }}`,
		`{{ turboStreamUpdate "m" }}{{ turboStreamEnd }}`,
		`{{ turboStreamRemove "m" }}{{ turboStreamEnd }}`,
		`{{ turboStreamBefore "m" }}{{ turboStreamEnd }}`,
		`{{ turboStreamAfter "m" }}{{ turboStreamEnd }}`,
		`{{ turboStreamRefresh }}{{ turboStreamEnd }}`,
		`{{ turboAttrTargets ".x" }}`,
		`{{ turboAttrRequestID "r-1" }}`,
		`{{ turboAttrMethodMorph }}`,
		`{{ turboStreamSourceSSE "s" }}`,
	}, "")

	tests := []struct {
		name         string
		wantKeys     []string
		templateSrc  string
		wantContains []string
	}{
		{
			name:         "every helper is registered and its closure executes",
			wantKeys:     wantKeys,
			templateSrc:  templateSrc,
			wantContains: []string{`<turbo-frame id="x">`, `inside`, `</turbo-frame>`},
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

			tmpl, err := template.New("all").Funcs(fm).Parse(tt.templateSrc)
			assert.NoError(t, err)

			var buf strings.Builder
			assert.NoError(t, tmpl.Execute(&buf, nil))
			for _, sub := range tt.wantContains {
				assert.Contains(t, buf.String(), sub)
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
