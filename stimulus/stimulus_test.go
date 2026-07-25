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
