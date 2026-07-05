package templatefunc

import (
	"fmt"
	"html/template"
)

// turboVersion pins the Turbo library version served by ScriptImport.
// Update in step with hotwire-go releases when validated against a new Turbo release.
const turboVersion = "8.0.23"

// ScriptImport renders the <script> tag that loads the Turbo runtime from a
// CDN. It must be placed inside <head>; without it, no Turbo behavior
// (Drive, Frames, Streams) is active on the page.
//
// The URL points to the ESM build served by jsDelivr, pinned to turboVersion.
// Applications that manage assets via a bundler or import map do not need
// this helper and can load Turbo through their own asset pipeline.
//
// Register via turbo.TemplateFuncMap and call from templates as:
//
//	{{ turboScriptImport }}
//
// Turbo installation reference:
// https://turbo.hotwired.dev/handbook/installing
func ScriptImport() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<script type="module" src="https://cdn.jsdelivr.net/npm/@hotwired/turbo@%s/dist/turbo.es2017-esm.js"></script>`,
		turboVersion,
	))
}
