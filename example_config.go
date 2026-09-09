package gavel

import _ "embed"

//go:generate go run gen_schema.go

// GavelConfigExample is the bundled annotated example for `.gavel.yaml`.
//
//go:embed gavel.yaml.example
var GavelConfigExample string
