package ui

import (
	"html/template"
	"strings"
)

// pageTemplateSource is the dashboard shell. It is an html/template rather than a
// concatenated string so the two injected values land in a parsed context instead
// of being spliced into markup: the bundled stylesheet in a CSS context and the
// build-info bootstrap in a script context. Splicing a JSON document between
// literal `<script>` … `</script>` tags is exactly the "unsafe quoting" pattern —
// a value carrying `</script>` (or a stray quote) closes the element early.
const pageTemplateSource = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>gavel · PR Dashboard</title>
    <link rel="icon" type="image/svg+xml" href="/favicon.svg">
    <link rel="apple-touch-icon" href="/brand/apple-touch-icon.png">
    <link rel="manifest" href="/manifest.webmanifest">
    <meta name="theme-color" content="#3578e5">
    <meta name="mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="default">
    <meta name="apple-mobile-web-app-title" content="gavel">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@400;500;600;700&family=Fira+Code:wght@400;500;600&display=swap" rel="stylesheet">
    <style>{{.BundleCSS}}</style>
    <style>
        @keyframes gavel-progress-slide {
            0%   { left: -35%; }
            100% { left: 100%; }
        }
        .gavel-progress-bar {
            animation: gavel-progress-slide 1.1s ease-in-out infinite;
        }
    </style>
</head>
<body class="bg-background text-foreground">
    <div id="root"></div>
    <script>window.__GAVEL__={{.Build}};</script>
    <script type="module" src="/_assets/prui.js"></script>
</body>
</html>`

var pageTemplate = template.Must(template.New("page").Parse(pageTemplateSource))

// pageData binds the shell's two dynamic values. BundleCSS is the embedded
// build artifact, marked as CSS so html/template emits it verbatim; Build is
// rendered by html/template's JS-context escaper, which serializes it as JSON
// and neutralizes any quote or `</script>` a build stamp might carry.
type pageData struct {
	BundleCSS template.CSS
	Build     BuildInfo
}

// pageHTML renders the dashboard shell. The template is static and validated at
// init by template.Must, and BuildInfo is a struct of plain strings, so an
// execution failure can only be a programming error in this file — it panics
// rather than serving a truncated page, matching assetsHandler's treatment of a
// broken embed.
func pageHTML() string {
	var buf strings.Builder
	if err := pageTemplate.Execute(&buf, pageData{BundleCSS: template.CSS(bundleCSS), Build: Build}); err != nil {
		panic("ui: render dashboard page: " + err.Error())
	}
	return buf.String()
}
