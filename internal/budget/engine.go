package budget

// Estimate is the price of one prompt on one tier.
type Estimate struct {
	Tokens     int
	Tier       string
	Surcharged bool
	USD        float64
}

// Calculator prices prompts against a Spec.
type Calculator struct {
	Spec Spec
}

// Estimate prices a prompt of the given size on the named tier.
//
// TODO(stage): apply spec.surcharge_multiplier at the long-context cliff.
func (c Calculator) Estimate(tokens int, tier string) (Estimate, error) {
	t, err := c.Spec.Tier(tier)
	if err != nil {
		return Estimate{}, err
	}
	usd := float64(tokens) / 1e6 * t.USDPerMillion
	return Estimate{Tokens: tokens, Tier: tier, USD: usd}, nil
}
