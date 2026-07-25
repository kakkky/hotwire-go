package stimulus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScriptLoad(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  Tag
	}{
		{
			name:  "no controllers renders only the importmap and empty module",
			paths: nil,
			want: `<script type="importmap">
{
  "imports": {
    "@hotwired/stimulus": "https://cdn.jsdelivr.net/npm/@hotwired/stimulus@3.2.2/dist/stimulus.js"
  }
}
</script>
<script type="module">
  import { Application } from "@hotwired/stimulus"

  window.Stimulus = Application.start()
</script>`,
		},
		{
			name:  "single controller with _controller suffix derives binding and identifier",
			paths: []string{"/js/hello_controller.js"},
			want: `<script type="importmap">
{
  "imports": {
    "@hotwired/stimulus": "https://cdn.jsdelivr.net/npm/@hotwired/stimulus@3.2.2/dist/stimulus.js"
  }
}
</script>
<script type="module">
  import { Application } from "@hotwired/stimulus"
  import HelloController from "/js/hello_controller.js"

  window.Stimulus = Application.start()
  Stimulus.register("hello", HelloController)
</script>`,
		},
		{
			name:  "multiple controllers are imported and registered in order",
			paths: []string{"/js/hello_controller.js", "/js/users_list_controller.js"},
			want: `<script type="importmap">
{
  "imports": {
    "@hotwired/stimulus": "https://cdn.jsdelivr.net/npm/@hotwired/stimulus@3.2.2/dist/stimulus.js"
  }
}
</script>
<script type="module">
  import { Application } from "@hotwired/stimulus"
  import HelloController from "/js/hello_controller.js"
  import UsersListController from "/js/users_list_controller.js"

  window.Stimulus = Application.start()
  Stimulus.register("hello", HelloController)
  Stimulus.register("users-list", UsersListController)
</script>`,
		},
		{
			name:  "basename without _controller suffix keeps the whole name as identifier",
			paths: []string{"/js/hello.js"},
			want: `<script type="importmap">
{
  "imports": {
    "@hotwired/stimulus": "https://cdn.jsdelivr.net/npm/@hotwired/stimulus@3.2.2/dist/stimulus.js"
  }
}
</script>
<script type="module">
  import { Application } from "@hotwired/stimulus"
  import Hello from "/js/hello.js"

  window.Stimulus = Application.start()
  Stimulus.register("hello", Hello)
</script>`,
		},
		{
			name:  "absolute URL is preserved and only the basename drives naming",
			paths: []string{"https://cdn.example.com/foo_bar_controller.js"},
			want: `<script type="importmap">
{
  "imports": {
    "@hotwired/stimulus": "https://cdn.jsdelivr.net/npm/@hotwired/stimulus@3.2.2/dist/stimulus.js"
  }
}
</script>
<script type="module">
  import { Application } from "@hotwired/stimulus"
  import FooBarController from "https://cdn.example.com/foo_bar_controller.js"

  window.Stimulus = Application.start()
  Stimulus.register("foo-bar", FooBarController)
</script>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ScriptLoad(tt.paths...))
		})
	}
}

func TestAttrControllers(t *testing.T) {
	tests := []struct {
		name        string
		identifiers []string
		want        Attrs
	}{
		{
			name:        "no identifiers renders an empty data-controller value",
			identifiers: nil,
			want:        Attrs{{Key: "data-controller", Value: ""}},
		},
		{
			name:        "single identifier",
			identifiers: []string{"hello"},
			want:        Attrs{{Key: "data-controller", Value: "hello"}},
		},
		{
			name:        "multiple identifiers are space-joined in order",
			identifiers: []string{"hello", "clipboard"},
			want:        Attrs{{Key: "data-controller", Value: "hello clipboard"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AttrControllers(tt.identifiers...))
		})
	}
}

func TestAttrAction(t *testing.T) {
	tests := []struct {
		name        string
		descriptors []string
		want        Attrs
	}{
		{
			name:        "no descriptors renders an empty data-action value",
			descriptors: nil,
			want:        Attrs{{Key: "data-action", Value: ""}},
		},
		{
			name:        "single descriptor",
			descriptors: []string{"click->gallery#next"},
			want:        Attrs{{Key: "data-action", Value: "click->gallery#next"}},
		},
		{
			name:        "multiple descriptors are space-joined in order",
			descriptors: []string{"focus->field#highlight", "input->search#update"},
			want:        Attrs{{Key: "data-action", Value: "focus->field#highlight input->search#update"}},
		},
		{
			name:        "KeyboardEvent filter is passed through verbatim",
			descriptors: []string{"keydown.esc->modal#close"},
			want:        Attrs{{Key: "data-action", Value: "keydown.esc->modal#close"}},
		},
		{
			name:        "global target @window is passed through verbatim",
			descriptors: []string{"resize@window->gallery#layout"},
			want:        Attrs{{Key: "data-action", Value: "resize@window->gallery#layout"}},
		},
		{
			name:        "action option on method is passed through verbatim",
			descriptors: []string{"scroll->gallery#layout:!passive"},
			want:        Attrs{{Key: "data-action", Value: "scroll->gallery#layout:!passive"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AttrAction(tt.descriptors...))
		})
	}
}

func TestAttrActionParam(t *testing.T) {
	tests := []struct {
		name       string
		controller string
		paramName  string
		paramValue string
		want       Attrs
	}{
		{
			name:       "number-like value",
			controller: "item",
			paramName:  "id",
			paramValue: "12345",
			want:       Attrs{{Key: "data-item-id-param", Value: "12345"}},
		},
		{
			name:       "string value",
			controller: "item",
			paramName:  "url",
			paramValue: "/votes",
			want:       Attrs{{Key: "data-item-url-param", Value: "/votes"}},
		},
		{
			name:       "object value as JSON string",
			controller: "item",
			paramName:  "payload",
			paramValue: `{"value":"1234567"}`,
			want:       Attrs{{Key: "data-item-payload-param", Value: `{"value":"1234567"}`}},
		},
		{
			name:       "boolean-like value",
			controller: "item",
			paramName:  "active",
			paramValue: "true",
			want:       Attrs{{Key: "data-item-active-param", Value: "true"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AttrActionParam(tt.controller, tt.paramName, tt.paramValue))
		})
	}
}
