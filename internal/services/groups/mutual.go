package groups

import "context"

// GroupSummary is a minimal group record for relationship APIs.
type GroupSummary struct {
	GroupID     string `json:"groupId"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	MemberCount int    `json:"memberCount"`
}

// GroupsForMember returns every group the member belongs to.
func (gs *GroupService) GroupsForMember(memberID string) []GroupSummary {
	ids, err := gs.store.ListGroupIDsForMember(context.Background(), memberID)
	if err != nil {
		return nil
	}
	var out []GroupSummary
	for _, groupID := range ids {
		g, err := gs.GetGroup(groupID)
		if err != nil {
			continue
		}
		out = append(out, GroupSummary{
			GroupID:     g.GroupID,
			Name:        g.Name,
			Type:        string(g.Type),
			MemberCount: g.CurrentMembers,
		})
	}
	return out
}

// MutualGroups returns groups where both DIDs are members.
func (gs *GroupService) MutualGroups(didA, didB string) []GroupSummary {
	if didA == "" || didB == "" || didA == didB {
		return nil
	}
	inB := make(map[string]struct{})
	for _, g := range gs.GroupsForMember(didB) {
		inB[g.GroupID] = struct{}{}
	}
	var mutual []GroupSummary
	for _, g := range gs.GroupsForMember(didA) {
		if _, ok := inB[g.GroupID]; ok {
			mutual = append(mutual, g)
		}
	}
	if mutual == nil {
		mutual = []GroupSummary{}
	}
	return mutual
}
