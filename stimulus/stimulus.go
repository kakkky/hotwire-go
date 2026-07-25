package stimulus

import (
	"fmt"
	"path"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{}
}

const stimulusVersion = "3.2.2"

const scriptImportMapHTML = `<script type="importmap">
{
  "imports": {
    "@hotwired/stimulus": "https://cdn.jsdelivr.net/npm/@hotwired/stimulus@` + stimulusVersion + `/dist/stimulus.js"
  }
}
</script>`

var scriptModuleHTMLTmpl = template.Must(template.New("module").Parse(
	`<script type="module">
  import { Application } from "@hotwired/stimulus"
{{- range . }}
  import {{ .VarName }} from "{{ .URL }}"
{{- end }}

  window.Stimulus = Application.start()
{{- range . }}
  Stimulus.register("{{ .Identifier }}", {{ .VarName }})
{{- end }}
</script>`))

type controllerEntry struct {
	URL        string
	VarName    string
	Identifier string
}

// ScriptLoad renders the pair of <script> tags that boot Stimulus: an
// importmap pinning @hotwired/stimulus to the jsDelivr build, and a module
// script that imports Application and each controller in controllerPaths,
// then calls Application.start() and app.register(...). It must be placed
// inside <head>.
//
// The controller path's basename drives naming: hello_controller.js becomes
// binding HelloController and identifier hello (trailing "_controller" is
// stripped, remaining underscores become dashes).
//
// Register via stimulus.TemplateFuncMap and call from templates as:
//
//	{{ stimulusScriptLoad "..." "..." }}
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the component syntax:
//
//	@stimulus.ScriptLoad("...", "...")
//
// Stimulus Handbook — Installing Stimulus for Development:
// https://stimulus.hotwired.dev/handbook/installing
func ScriptLoad(controllerPaths ...string) Tag {
	titler := cases.Title(language.English)

	entries := make([]controllerEntry, 0, len(controllerPaths))
	for _, c := range controllerPaths {
		base := path.Base(c)
		snakeCaseName := strings.TrimSuffix(base, path.Ext(base))

		varParts := strings.Split(snakeCaseName, "_")
		for i, vp := range varParts {
			if vp == "" {
				continue
			}
			varParts[i] = titler.String(vp)
		}
		varName := strings.Join(varParts, "")

		identifier := strings.ReplaceAll(strings.TrimSuffix(snakeCaseName, "_controller"), "_", "-")
		entries = append(entries, controllerEntry{
			URL:        c,
			VarName:    varName,
			Identifier: identifier,
		})
	}

	var b strings.Builder
	b.WriteString(scriptImportMapHTML)
	b.WriteByte('\n')
	if err := scriptModuleHTMLTmpl.Execute(&b, entries); err != nil {
		panic(err)
	}
	return Tag(b.String())
}

// AttrControllers renders data-controller="{id1} {id2} ..." on any
// element.
//
// Stimulus scans the DOM for this attribute and instantiates each named
// controller with the element as its scope. Multiple space-separated
// identifiers attach multiple controllers to the same element, and the
// same controller may appear on any number of elements. Identifiers
// must match those registered by ScriptLoad (or Application.register).
//
// Each identifier is HTML-escaped before insertion; pass plain
// identifier strings.
//
// Register via stimulus.TemplateFuncMap and call from templates as:
//
//	<div {{ stimulusAttrControllers "..." "..." }}>...</div>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<div { stimulus.AttrControllers("...", "...")... }>...</div>
//
// Stimulus Reference — Controllers:
// https://stimulus.hotwired.dev/reference/controllers
func AttrControllers(identifiers ...string) Attrs {
	return Attrs{{Key: "data-controller", Value: strings.Join(identifiers, " ")}}
}

// AttrAction renders data-action="{d1} {d2} ..." on any element,
// where each di is a Stimulus action descriptor.
//
// A descriptor has the form event->controller#method:
//
//   - event: DOM event name (for example click, submit, input, change,
//     focus, blur, keydown, keyup, scroll).
//   - controller: Stimulus identifier registered via ScriptLoad (or
//     Application.register).
//   - method: controller method to call, camelCase.
//
// # KeyboardEvent filter
//
// For keyboard events, append a filter to the event name with a dot
// (for example keydown.esc, keydown.ctrl+a). The method fires only
// when the matching key (or combination) is pressed.
//
// Supported filters: enter, tab, esc, space, up, down, left, right,
// home, end, page_up, page_down, [a-z], [0-9].
//
// Supported modifiers: alt (option on macOS), ctrl, meta (Command on
// macOS), shift. Combine with + (for example keydown.ctrl+a).
//
// # Global events
//
// Append @window or @document to the event name to listen on window
// or document instead of the element itself (for example
// resize@window, visibilitychange@document).
//
// # Action options
//
// Append options to the method with a colon (for example
// scroll->foo#layout:!passive). DOM listener options:
//
//   - :capture   → { capture: true }
//   - :once      → { once: true }
//   - :passive   → { passive: true }
//   - :!passive  → { passive: false }
//
// Stimulus-specific options:
//
//   - :stop     call stopPropagation() before the method
//   - :prevent  call preventDefault() before the method
//   - :self     fire only when the event target is the element itself
//
// # Multiple descriptors
//
// Pass multiple descriptors to attach several actions to the same
// element; they are space-joined into a single data-action value and
// invoked by Stimulus left-to-right.
//
// Register via stimulus.TemplateFuncMap and call from templates as:
//
//	<button {{ stimulusAttrAction "..." }}>...</button>
//	<button {{ stimulusAttrAction "..." "..." }}>...</button>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<button { stimulus.AttrAction("...")... }>...</button>
//	<button { stimulus.AttrAction("...", "...")... }>...</button>
//
// Stimulus Reference — Actions:
// https://stimulus.hotwired.dev/reference/actions
func AttrAction(descriptors ...string) Attrs {
	return Attrs{{Key: "data-action", Value: strings.Join(descriptors, " ")}}
}

// AttrActionParam renders data-{controller}-{paramName}-param="{paramValue}"
// on an element. The attribute must sit on the same element as the
// action that will read it; Stimulus exposes the value on the action
// method's event.params.
//
// Stimulus infers the value type (Number, String, Object, Boolean)
// from the string form, so this helper always writes the value as-is.
// For example "12345" is read as a Number, `{"key":"value"}` as an
// Object, "true" as a Boolean.
//
// Register via stimulus.TemplateFuncMap and call from templates as:
//
//	<button {{ stimulusAttrAction "..." }}
//	        {{ stimulusAttrActionParam "..." "..." "..." }}>...</button>
//
// Alternatively, call it directly from an a-h/templ template
// (https://github.com/a-h/templ) using the spread attributes syntax:
//
//	<button { stimulus.AttrAction("...")... }
//	        { stimulus.AttrActionParam("...", "...", "...")... }>...</button>
//
// Stimulus Reference — Action Parameters:
// https://stimulus.hotwired.dev/reference/actions#action-parameters
func AttrActionParam(controller, paramName, paramValue string) Attrs {
	return Attrs{{Key: fmt.Sprintf("data-%s-%s-param", controller, paramName), Value: paramValue}}
}
