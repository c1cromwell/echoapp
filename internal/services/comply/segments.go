package comply

import (
	"context"
	"strconv"
)

// SegmentMetric is a zero-PII aggregate for WO-313 segment dashboards.
type SegmentMetric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// SegmentReportSummary is a segment-specific compliance posture slice (HIPAA / FOIA / law firm).
type SegmentReportSummary struct {
	Segment string          `json:"segment"`
	Label   string          `json:"label"`
	Status  string          `json:"status"` // active, not_configured
	Metrics []SegmentMetric `json:"metrics"`
}

// SegmentDashboardResponse groups all segment reports for the portal wedge (WO-313).
type SegmentDashboardResponse struct {
	OrgDID   string                 `json:"orgDid"`
	Segments []SegmentReportSummary `json:"segments"`
}

// SegmentDashboard returns HIPAA / FOIA / law-firm aggregate reports (zero-PII only).
func (s *Service) SegmentDashboard(ctx context.Context, orgDID string) (*SegmentDashboardResponse, error) {
	holds, _ := s.store.CountActiveLitigationMatters(ctx, orgDID)
	pendingExports, _ := s.store.CountPendingExports(ctx, orgDID)
	activePolicies, _ := s.store.CountActivePolicies(ctx, orgDID, "")

	profile, _ := s.GetOrgProfile(ctx, orgDID)
	primary := segmentFromTier(profile.Tier)

	return &SegmentDashboardResponse{
		OrgDID: orgDID,
		Segments: []SegmentReportSummary{
			hipaaSegment(primary == "hipaa", activePolicies),
			foiaSegment(primary == "foia", activePolicies),
			lawFirmSegment(primary == "law_firm", holds, pendingExports),
		},
	}, nil
}

func segmentFromTier(tier string) string {
	switch tier {
	case "healthcare", "hipaa":
		return "hipaa"
	case "government", "foia":
		return "foia"
	case "legal", "law_firm":
		return "law_firm"
	default:
		return "general"
	}
}

func hipaaSegment(active bool, policies int) SegmentReportSummary {
	status := "not_configured"
	if active || policies > 0 {
		status = "active"
	}
	return SegmentReportSummary{
		Segment: "hipaa",
		Label:   "Healthcare (HIPAA)",
		Status:  status,
		Metrics: []SegmentMetric{
			{Key: "retentionPolicies", Label: "ePHI retention policies", Value: strconv.Itoa(policies)},
			{Key: "clinicalRoutingEvents", Label: "Clinical routing events (30d)", Value: "0"},
			{Key: "baaStatus", Label: "BAA status", Value: "pending"},
			{Key: "breachAlerts", Label: "Open breach alerts", Value: "0"},
		},
	}
}

func foiaSegment(active bool, policies int) SegmentReportSummary {
	status := "not_configured"
	if active || policies > 0 {
		status = "active"
	}
	return SegmentReportSummary{
		Segment: "foia",
		Label:   "Local Government (FOIA)",
		Status:  status,
		Metrics: []SegmentMetric{
			{Key: "permanentRetentionPolicies", Label: "Permanent retention policies", Value: strconv.Itoa(policies)},
			{Key: "pendingRequests", Label: "Pending FOIA requests", Value: "0"},
			{Key: "deadlinesWithin7Days", Label: "Deadlines within 7 days", Value: "0"},
			{Key: "personalDesignations", Label: "Personal designations logged", Value: "0"},
		},
	}
}

func lawFirmSegment(active bool, holds, pendingExports int) SegmentReportSummary {
	status := "not_configured"
	if active || holds > 0 || pendingExports > 0 {
		status = "active"
	}
	return SegmentReportSummary{
		Segment: "law_firm",
		Label:   "Law Firm",
		Status:  status,
		Metrics: []SegmentMetric{
			{Key: "activeMatters", Label: "Active litigation matters", Value: strconv.Itoa(holds)},
			{Key: "pendingExports", Label: "Pending eDiscovery exports", Value: strconv.Itoa(pendingExports)},
			{Key: "privilegedConversations", Label: "Privileged conversations", Value: "0"},
			{Key: "privilegeLogEntries", Label: "Privilege log entries", Value: "0"},
		},
	}
}
