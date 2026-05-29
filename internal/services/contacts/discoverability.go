package contacts

// minTierForDefaultPhoneDiscovery is the trust tier at which users are
// discoverable via PSI without an explicit opt-in (WO-220).
const minTierForDefaultPhoneDiscovery = 3

// IsPhoneDiscoverable reports whether a user's phone may appear in the PSI index.
// Tier 3+ users are discoverable by default; lower tiers require explicit opt-in.
func IsPhoneDiscoverable(tier int, optIn *bool) bool {
	if optIn != nil {
		return *optIn
	}
	return tier >= minTierForDefaultPhoneDiscovery
}
