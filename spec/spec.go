// Package spec embeds the tokburn pricing spec, the single source of truth for prices.
package spec

import _ "embed"

//go:embed pricing.yaml
var Pricing []byte
