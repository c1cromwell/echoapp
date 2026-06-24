package bots

// DefaultCatalog returns curated stub bots for Stage 4 marketplace MVP.
func DefaultCatalog() []Manifest {
	return []Manifest{
		{
			BotDID:      "did:key:z6MkechoReminderBot00000000000000000001",
			Name:        "Echo Reminder",
			Description: "Schedule gentle nudges in your chats — on-device rules only.",
			Version:     "0.1.0",
			RequiredPermissions: []Permission{
				PermSendMessage,
				PermReadMessages,
			},
			TrustScore: 82,
		},
		{
			BotDID:      "did:key:z6MkechoTranslateBot00000000000000000002",
			Name:        "Translate Helper",
			Description: "Suggests on-device translations when you long-press a message.",
			Version:     "0.1.0",
			RequiredPermissions: []Permission{
				PermReadMessages,
			},
			TrustScore: 88,
		},
		{
			BotDID:      "did:key:z6MkechoTipJarBot000000000000000000003",
			Name:        "Tip Jar",
			Description: "Request ECHO tips in chat with explicit per-payment approval.",
			Version:     "0.1.0",
			RequiredPermissions: []Permission{
				PermSendMessage,
				PermRequestPayment,
			},
			TrustScore: 75,
		},
	}
}

// LookupManifest returns a catalog entry by bot DID.
func LookupManifest(botDID string) (Manifest, bool) {
	for _, m := range DefaultCatalog() {
		if m.BotDID == botDID {
			return m, true
		}
	}
	return Manifest{}, false
}
