package turbo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriveTagHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  Tag
		want Tag
	}{
		{
			name: "MetaCacheControlNoPreview",
			got:  MetaCacheControlNoPreview(),
			want: `<meta name="turbo-cache-control" content="no-preview">`,
		},
		{
			name: "MetaCacheControlNoCache",
			got:  MetaCacheControlNoCache(),
			want: `<meta name="turbo-cache-control" content="no-cache">`,
		},
		{
			name: "MetaViewTransition",
			got:  MetaViewTransition(),
			want: `<meta name="view-transition" content="same-origin">`,
		},
		{
			name: "MetaRefreshMethodMorph",
			got:  MetaRefreshMethodMorph(),
			want: `<meta name="turbo-refresh-method" content="morph">`,
		},
		{
			name: "MetaRefreshScrollPreserve",
			got:  MetaRefreshScrollPreserve(),
			want: `<meta name="turbo-refresh-scroll" content="preserve">`,
		},
		{
			name: "MetaDisablePrefetch",
			got:  MetaDisablePrefetch(),
			want: `<meta name="turbo-prefetch" content="false">`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestMetaRoot(t *testing.T) {
	tests := []struct {
		name string
		path string
		want Tag
	}{
		{
			name: "plain path",
			path: "/app",
			want: `<meta name="turbo-root" content="/app">`,
		},
		{
			name: "path with special characters is HTML-escaped",
			path: `/app"<script>`,
			want: `<meta name="turbo-root" content="/app&#34;&lt;script&gt;">`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, MetaRoot(tt.path))
		})
	}
}

func TestDriveAttrHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  Attrs
		want Attrs
	}{
		{
			name: "AttrTrackReload",
			got:  AttrTrackReload(),
			want: Attrs{{Key: "data-turbo-track", Value: "reload"}},
		},
		{
			name: "AttrTrackDynamic",
			got:  AttrTrackDynamic(),
			want: Attrs{{Key: "data-turbo-track", Value: "dynamic"}},
		},
		{
			name: "AttrDisableTurbo",
			got:  AttrDisableTurbo(),
			want: Attrs{{Key: "data-turbo", Value: "false"}},
		},
		{
			name: "AttrEnableTurbo",
			got:  AttrEnableTurbo(),
			want: Attrs{{Key: "data-turbo", Value: "true"}},
		},
		{
			name: "AttrPreload",
			got:  AttrPreload(),
			want: Attrs{{Key: "data-turbo-preload", Value: true}},
		},
		{
			name: "AttrDisablePrefetch",
			got:  AttrDisablePrefetch(),
			want: Attrs{{Key: "data-turbo-prefetch", Value: "false"}},
		},
		{
			name: "AttrPermanent",
			got:  AttrPermanent(),
			want: Attrs{{Key: "data-turbo-permanent", Value: true}},
		},
		{
			name: "AttrTemporary",
			got:  AttrTemporary(),
			want: Attrs{{Key: "data-turbo-temporary", Value: true}},
		},
		{
			name: "AttrDisableEval",
			got:  AttrDisableEval(),
			want: Attrs{{Key: "data-turbo-eval", Value: "false"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
