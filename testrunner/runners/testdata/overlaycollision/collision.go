// Package overlaycollision declares, at package level, every identifier the
// injected Ginkgo reporter would otherwise bring into file scope.
package overlaycollision

type Report struct{}

type Context struct{}

type Labels []string

var (
	types  = "types"
	json   = "json"
	fmt    = "fmt"
	os     = "os"
	base64 = "base64"
)

func Names() []string { return []string{types, json, fmt, os, base64} }
