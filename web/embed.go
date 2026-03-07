// Package web uses Go's embed package to bundle static files (HTML, CSS, JS)
// directly into the compiled binary. This means you don't need to ship a
// separate "web" folder alongside the binary — everything is self-contained.
package web

import "embed"

// StaticFS embeds the index.html file at compile time.
// The //go:embed directive tells the Go compiler to include the file's
// contents in the binary. At runtime, StaticFS acts like a read-only filesystem.
//
//go:embed index.html
var StaticFS embed.FS
