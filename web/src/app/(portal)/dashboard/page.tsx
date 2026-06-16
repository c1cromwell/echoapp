import Link from "next/link";
import { getActiveOrg } from "@/lib/auth/org";
import {
  getAuditReport,
  getComplyDashboard,
} from "@/lib/api/comply";

export default async function DashboardPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  const [summary, audit] = await Promise.all([
    getComplyDashboard(org.orgDID).catch(() => null),
    getAuditReport(org.orgDID).catch(() => null),
  ]);

  const cards = [
    { label: "Digital-Evidence coverage", value: summary?.deCoverageRate ?? "—" },
    { label: "Active retention policies", value: summary?.activeRetentionPolicies ?? "—" },
    { label: "Litigation holds", value: summary?.litigationHolds ?? "—" },
    { label: "Pending exports", value: summary?.pendingExports ?? "—" },
    {
      label: "Metagraph anchor health",
      value: summary?.anchorHealth ?? "—",
      tone:
        summary?.anchorHealth === "healthy"
          ? "text-emerald-400"
          : summary?.anchorHealth === "degraded"
            ? "text-amber-400"
            : "text-red-400",
    },
  ];

  return (
    <div>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-white">Compliance posture</h1>
          <p className="mt-1 text-sm text-glacial-muted">
            {org.orgName} — zero-PII view (hashes, CIDs, and aggregate metrics only).
          </p>
        </div>
        <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
          <a
            href="/api/comply/audit"
            className="rounded-lg border border-glacial-border px-3 py-2 text-center text-xs text-white hover:bg-glacial-surface"
          >
            Export audit PDF
          </a>
          <Link
            href="/retention"
            className="rounded-lg bg-white/10 px-3 py-2 text-center text-xs text-white hover:bg-white/15"
          >
            Manage policies
          </Link>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-3 xs:grid-cols-2 sm:gap-4 lg:grid-cols-3">
        {cards.map((c) => (
          <div
            key={c.label}
            className="rounded-2xl border border-glacial-border bg-glacial-surface p-5"
          >
            <p className="text-xs text-glacial-muted">{c.label}</p>
            <p className={`mt-2 text-2xl font-semibold ${c.tone ?? "text-white"}`}>
              {String(c.value)}
            </p>
          </div>
        ))}
      </div>

      {audit && (
        <section className="mt-8 rounded-2xl border border-glacial-border bg-glacial-surface p-5">
          <h2 className="text-sm font-semibold text-white">Audit snapshot</h2>
          <p className="mt-2 text-xs text-glacial-muted">{audit.verificationNotice}</p>
          <dl className="mt-4 grid grid-cols-2 gap-3 text-xs text-glacial-muted">
            <div>
              <dt>Report generated</dt>
              <dd className="text-white">
                {audit.generatedAt
                  ? new Date(audit.generatedAt).toLocaleString()
                  : "—"}
              </dd>
            </div>
            <div>
              <dt>Events in range</dt>
              <dd className="text-white">{audit.activeRetentionPolicies} policies tracked</dd>
            </div>
          </dl>
        </section>
      )}

      {!summary && (
        <p className="mt-6 text-xs text-glacial-muted">
          Comply API not reachable — set <code>COMPLY_API_BASE_URL</code> (default{" "}
          <code>http://localhost:8011</code>).
        </p>
      )}
    </div>
  );
}
