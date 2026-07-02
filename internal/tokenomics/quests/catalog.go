package quests

// Tier distinguishes starter vs advanced quest catalogs.
type Tier string

const (
	TierStarter  Tier = "starter"
	TierAdvanced Tier = "advanced"
)

// Definition is a quest catalog entry (WO-271).
type Definition struct {
	ID            string `json:"questId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Action        string `json:"action"`
	RequiredCount int    `json:"requiredCount"`
	RewardECHO    int64  `json:"reward_echo"` // datum
	Badge         string `json:"badge"`
	Tier          Tier   `json:"tier"`
	MinTrustTier  int    `json:"minTrustTier,omitempty"`
}

// Completion tracks per-DID quest progress on Data L1 mirror.
type Completion struct {
	DID           string `json:"did"`
	QuestID       string `json:"quest_id"`
	CompletedAt   string `json:"completed_at,omitempty"`
	RewardClaimed bool   `json:"reward_claimed"`
	TxHash        string `json:"tx_hash,omitempty"`
}

// CatalogEntry is returned by GET /gamification/quests.
type CatalogEntry struct {
	Definition
	CompletedAt   string `json:"completedAt,omitempty"`
	RewardClaimed bool   `json:"rewardClaimed"`
	Progress      int    `json:"progress"`
}

// StarterQuests is the Phase 2 launch catalog.
var StarterQuests = []Definition{
	{ID: "identity_builder", Title: "Identity Builder", Description: "Complete identity verification", Action: "identity_verification", RequiredCount: 1, RewardECHO: 20 * 100_000_000, Badge: "Verified", Tier: TierStarter},
	{ID: "community_joiner", Title: "Community Joiner", Description: "Join or create a group with 5+ members", Action: "group_join", RequiredCount: 1, RewardECHO: 10 * 100_000_000, Badge: "Group Member", Tier: TierStarter},
	{ID: "trusted_messenger", Title: "Trusted Messenger", Description: "Reach Trust Tier 3", Action: "trust_tier", RequiredCount: 3, RewardECHO: 50 * 100_000_000, Badge: "Trusted", Tier: TierStarter, MinTrustTier: 3},
	{ID: "stack_and_earn", Title: "Stack and Earn", Description: "Stake ECHO for the first time", Action: "stake", RequiredCount: 1, RewardECHO: 15 * 100_000_000, Badge: "Staker", Tier: TierStarter},
	{ID: "invite_and_grow", Title: "Invite and Grow", Description: "Complete first successful referral", Action: "referral", RequiredCount: 1, RewardECHO: 25 * 100_000_000, Badge: "Connector", Tier: TierStarter},
	{ID: "vault_keeper", Title: "Vault Keeper", Description: "Send a disappearing message", Action: "disappearing_message", RequiredCount: 1, RewardECHO: 5 * 100_000_000, Badge: "Ghost", Tier: TierStarter},
	{ID: "private_circle", Title: "Private Circle", Description: "Activate a Hidden Folder", Action: "hidden_folder", RequiredCount: 1, RewardECHO: 10 * 100_000_000, Badge: "Vault", Tier: TierStarter},
	{ID: "vip_experience", Title: "VIP Experience", Description: "Upgrade to VIP for first month", Action: "vip_subscription", RequiredCount: 1, RewardECHO: 50 * 100_000_000, Badge: "VIP", Tier: TierStarter},
	{ID: "governance_debut", Title: "Governance Debut", Description: "Cast first governance vote", Action: "governance_vote", RequiredCount: 1, RewardECHO: 25 * 100_000_000, Badge: "Voter", Tier: TierStarter, MinTrustTier: 2},
}

// AdvancedQuests is the Phase 3+ catalog.
var AdvancedQuests = []Definition{
	{ID: "network_validator", Title: "Network Validator", Description: "Delegate to a validator for 30 days", Action: "delegate_30d", RequiredCount: 1, RewardECHO: 100 * 100_000_000, Badge: "Delegator", Tier: TierAdvanced},
	{ID: "whale_staker", Title: "Whale Staker", Description: "Stake Platinum tier (365 days)", Action: "stake_platinum", RequiredCount: 1, RewardECHO: 500 * 100_000_000, Badge: "Whale Staker", Tier: TierAdvanced},
	{ID: "network_builder", Title: "Network Builder", Description: "Refer 10 active users", Action: "referral_10", RequiredCount: 10, RewardECHO: 500 * 100_000_000, Badge: "Builder", Tier: TierAdvanced},
	{ID: "bot_creator", Title: "Bot Creator", Description: "Publish a bot to the marketplace", Action: "bot_publish", RequiredCount: 1, RewardECHO: 200 * 100_000_000, Badge: "Bot Creator", Tier: TierAdvanced},
}

// All returns the full quest catalog.
func All() []Definition {
	out := make([]Definition, 0, len(StarterQuests)+len(AdvancedQuests))
	out = append(out, StarterQuests...)
	out = append(out, AdvancedQuests...)
	return out
}

// ByID looks up a quest definition.
func ByID(id string) (Definition, bool) {
	for _, q := range All() {
		if q.ID == id {
			return q, true
		}
	}
	return Definition{}, false
}
