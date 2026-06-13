import { getActiveOrg } from "@/lib/auth/org";
import { getComplyDashboard } from "@/lib/api/comply";

/**
 * Reporting / analytics / auditing dashboard — WO-313 lands the full widget set here.
 * This is the placeholder shell (WO-309): it proves the org-scoped server fetch path
 * to the Go Comply API. All values are zero-PII: hashes/CIDs/aggregate counts only.
 */
export default async function DashboardPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  const summary = await getComplyDashboard(org.orgDID).catch(() => null);

  const cards = [
    { label: "Digital-Evidence coverage", value: summary?.deCoverageRate ?? "—" },
    { label: "Active retention policies", value: summary?.activeRetentionPolicies ?? "—" },
    { label: "Litigation holds", value: summary?.litigationHolds ?? "—" },
    { label: "Pending exports", value: summary?.pendingExports ?? "—" },
    { label: "Metagraph anchor health", value: summary?.anchorHealth ?? "—" },
  ];

  return (
    <div>
      <h1 className="text-2xl font-semibold text-white">Compliance posture</h1>
      <p className="mt-1 text-sm text-glacial-muted">
        {org.orgName} — zero-PII view (hashes, CIDs, and aggregate metrics only).
      </p>

      <div className="mt-6 grid grid-cols-2 gap-4 lg:grid-cols-3">
        {cards.map((c) => (
          <div
            key={c.label}
            className="rounded-2xl border border-glacial-border bg-glacial-surface p-5"
          >
            <p className="text-xs text-glacial-muted">{c.label}</p>
            <p className="mt-2 text-2xl font-semibold text-white">{String(c.value)}</p>
          </div>
        ))}
      </div>

      {!summary && (
        <p className="mt-6 text-xs text-glacial-muted">
          Comply API not reachable (set <code>COMPLY_API_BASE_URL</code>). Widgets populate once
          the WO-252 <code>/comply/dashboard</code> endpoint is live.
        </p>
      )}
    </div>
  );
}
