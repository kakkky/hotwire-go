package turbo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFrameRequest(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{
			name:   "Turbo-Frame header present",
			header: "posts",
			want:   true,
		},
		{
			name:   "Turbo-Frame header absent",
			header: "",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("Turbo-Frame", tt.header)
			}
			assert.Equal(t, tt.want, IsFrameRequest(r))
		})
	}
}

func TestFrameID(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "header value is returned as-is",
			header: "posts-list",
			want:   "posts-list",
		},
		{
			name:   "missing header returns empty string",
			header: "",
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("Turbo-Frame", tt.header)
			}
			assert.Equal(t, tt.want, FrameID(r))
		})
	}
}

func TestFrame(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		extra []Attrs
		want  Elm
	}{
		{
			name: "id only",
			id:   "posts",
			want: Elm{
				Tag:   Tag("turbo-frame"),
				Attrs: Attrs{{Key: "id", Value: "posts"}},
			},
		},
		{
			name:  "id with extra attrs",
			id:    "posts",
			extra: []Attrs{AttrSrc("/posts"), AttrLoadingLazy()},
			want: Elm{
				Tag: Tag("turbo-frame"),
				Attrs: Attrs{
					{Key: "id", Value: "posts"},
					{Key: "src", Value: "/posts"},
					{Key: "loading", Value: "lazy"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Frame(tt.id, tt.extra...))
		})
	}
}

func TestFrameAttrHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  Attrs
		want Attrs
	}{
		{
			name: "AttrSrc",
			got:  AttrSrc("/posts/42"),
			want: Attrs{{Key: "src", Value: "/posts/42"}},
		},
		{
			name: "AttrLoadingLazy",
			got:  AttrLoadingLazy(),
			want: Attrs{{Key: "loading", Value: "lazy"}},
		},
		{
			name: "AttrLoadingEager",
			got:  AttrLoadingEager(),
			want: Attrs{{Key: "loading", Value: "eager"}},
		},
		{
			name: "AttrDisabled",
			got:  AttrDisabled(),
			want: Attrs{{Key: "disabled", Value: true}},
		},
		{
			name: "AttrTarget",
			got:  AttrTarget("_top"),
			want: Attrs{{Key: "target", Value: "_top"}},
		},
		{
			name: "AttrFrame",
			got:  AttrFrame("_top"),
			want: Attrs{{Key: "data-turbo-frame", Value: "_top"}},
		},
		{
			name: "AttrRecurse",
			got:  AttrRecurse("nested"),
			want: Attrs{{Key: "recurse", Value: "nested"}},
		},
		{
			name: "AttrAutoscroll",
			got:  AttrAutoscroll(),
			want: Attrs{{Key: "autoscroll", Value: true}},
		},
		{
			name: "AttrAutoscrollBlockStart",
			got:  AttrAutoscrollBlockStart(),
			want: Attrs{{Key: "data-autoscroll-block", Value: "start"}},
		},
		{
			name: "AttrAutoscrollBlockCenter",
			got:  AttrAutoscrollBlockCenter(),
			want: Attrs{{Key: "data-autoscroll-block", Value: "center"}},
		},
		{
			name: "AttrAutoscrollBlockEnd",
			got:  AttrAutoscrollBlockEnd(),
			want: Attrs{{Key: "data-autoscroll-block", Value: "end"}},
		},
		{
			name: "AttrAutoscrollBlockNearest",
			got:  AttrAutoscrollBlockNearest(),
			want: Attrs{{Key: "data-autoscroll-block", Value: "nearest"}},
		},
		{
			name: "AttrAutoscrollBehaviorAuto",
			got:  AttrAutoscrollBehaviorAuto(),
			want: Attrs{{Key: "data-autoscroll-behavior", Value: "auto"}},
		},
		{
			name: "AttrAutoscrollBehaviorSmooth",
			got:  AttrAutoscrollBehaviorSmooth(),
			want: Attrs{{Key: "data-autoscroll-behavior", Value: "smooth"}},
		},
		{
			name: "AttrRefreshMorph",
			got:  AttrRefreshMorph(),
			want: Attrs{{Key: "refresh", Value: "morph"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}
