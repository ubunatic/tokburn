// Package opus embeds the owner-selected terminal screen designs.
package opus

import _ "embed"

//go:embed title.ansi
var Title string

//go:embed game.ansi
var Game string
