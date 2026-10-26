// Package flames embeds the owner-selected title animation.
package flames

import "embed"

//go:embed frame-*.ansi
var Frames embed.FS
