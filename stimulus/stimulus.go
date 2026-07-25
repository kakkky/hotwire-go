package stimulus

import (
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
