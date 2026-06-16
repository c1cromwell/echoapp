import "server-only";

const BASE = process.env.COMPLY_API_BASE_URL;
const TOKEN = process.env.COMPLY_API_SERVICE_TOKEN;

async function complyFetch<T>(
  path: string,
  orgDID: string,
  init?: RequestInit,
): Promise<T> {
  if (!BASE) throw new Error("COMPLY_API_BASE_URL not configured");

  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      Authorization: `Bearer ${TOKEN ?? ""}`,
      "X-Org-DID": orgDID,
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Comply API ${path} -> ${res.status}`);
  }
  return (await res.json()) as T;
}

/** WO-252 / WO-313 dashboard summary — zero-PII aggregates only. */
export interface ComplyDashboardSummary {
  deCoverageRate: string;
  activeRetentionPolicies: number;
  litigationHolds: number;
  pendingExports: number;
  anchorHealth: "healthy" | "degraded" | "down";
}

export interface RetentionPolicy {
  id: string;
  orgDid: string;
  policyType: "permanent" | "time_limited" | "litigation_hold";
  conversationId?: string;
  scopeLabel?: string;
  active: boolean;
}

export interface LitigationMatter {
  matterId: string;
  orgDid: string;
  scopeLabel?: string;
  status: "active" | "released";
  custodianCount: number;
  dataL1Ref?: string;
}

export interface EDiscoveryExport {
  exportId: string;
  matterId: string;
  status: "pending" | "processing" | "ready" | "delivered" | "failed";
  messageCount: number;
  queryHash: string;
  dataL1Ref?: string;
}

export interface AuditReport {
  orgDid: string;
  generatedAt: string;
  activeRetentionPolicies: number;
  activeLitigationHolds: number;
  pendingExports: number;
  anchorHealth: string;
  verificationNotice: string;
}

export function getComplyDashboard(orgDID: string) {
  return complyFetch<ComplyDashboardSummary>("/comply/dashboard", orgDID);
}

export function listRetentionPolicies(orgDID: string) {
  return complyFetch<{ policies: RetentionPolicy[] }>("/comply/retention/policy", orgDID);
}

export function createRetentionPolicy(
  orgDID: string,
  body: { policy_type: string; conversation_id?: string; scope_label?: string },
) {
  return complyFetch<RetentionPolicy>("/comply/retention/policy", orgDID, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function activateLitigationHold(
  orgDID: string,
  body: { matterId: string; custodianDids: string[]; scope?: string },
) {
  return complyFetch<LitigationMatter>("/comply/litigation/hold", orgDID, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function listEDiscoveryExports(orgDID: string) {
  return complyFetch<{ exports: EDiscoveryExport[] }>("/comply/ediscovery/export", orgDID);
}

export function requestEDiscoveryExport(
  orgDID: string,
  body: { matterId: string; custodianSet?: string[]; dateFrom?: string; dateTo?: string },
) {
  return complyFetch<EDiscoveryExport>("/comply/ediscovery/export", orgDID, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function getAuditReport(orgDID: string, from?: string, to?: string) {
  const q = new URLSearchParams({ format: "json" });
  if (from) q.set("from", from);
  if (to) q.set("to", to);
  return complyFetch<AuditReport>(`/comply/audit/report?${q}`, orgDID);
}

export function auditReportDownloadURL(orgDID: string) {
  const q = new URLSearchParams({ format: "pdf" });
  return `${BASE}/comply/audit/report?${q}`;
}
