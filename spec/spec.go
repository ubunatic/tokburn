// Package spec embeds the tokburn pricing and task specs.
package spec

import _ "embed"

//go:embed pricing.yaml
var Pricing []byte

// Tasks contains the small task deck with provisional tiers.
//
//go:embed tasks.yaml
var Tasks []byte
