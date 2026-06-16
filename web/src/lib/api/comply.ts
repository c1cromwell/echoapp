import "server-only";

import type { components } from "@/lib/api/schema";

const BASE = process.env.COMPLY_API_BASE_URL;
const TOKEN = process.env.COMPLY_API_SERVICE_TOKEN;

export type ComplyDashboardSummary = components["schemas"]["ComplyDashboardSummary"];
export type RetentionPolicy = components["schemas"]["ComplyRetentionPolicy"];
export type LitigationMatter = components["schemas"]["ComplyLitigationMatter"];
export type EDiscoveryExport = components["schemas"]["ComplyEDiscoveryExport"];
export type AuditReport = components["schemas"]["ComplyAuditReport"];
export type ComplyOrgProfile = components["schemas"]["ComplyOrgProfile"];
export type RetentionPolicyCreate = components["schemas"]["ComplyRetentionPolicyCreate"];
export type LitigationHoldRequest = components["schemas"]["ComplyLitigationHoldRequest"];
export type EDiscoveryExportRequest = components["schemas"]["ComplyEDiscoveryExportRequest"];

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

export function getComplyDashboard(orgDID: string) {
  return complyFetch<ComplyDashboardSummary>("/comply/dashboard", orgDID);
}

export function getOrgProfile(orgDID: string) {
  return complyFetch<ComplyOrgProfile>("/comply/org/profile", orgDID);
}

export function listLitigationHolds(orgDID: string, status: "active" | "all" = "active") {
  const q = status === "all" ? "?status=all" : "";
  return complyFetch<{ matters: LitigationMatter[] }>(`/comply/litigation/hold${q}`, orgDID);
}

export function releaseLitigationHold(orgDID: string, matterId: string) {
  return complyFetch<LitigationMatter>(`/comply/litigation/hold/${matterId}/release`, orgDID, {
    method: "PUT",
  });
}

export function listRetentionPolicies(orgDID: string) {
  return complyFetch<{ policies: RetentionPolicy[] }>("/comply/retention/policy", orgDID);
}

export function createRetentionPolicy(orgDID: string, body: RetentionPolicyCreate) {
  return complyFetch<RetentionPolicy>("/comply/retention/policy", orgDID, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function activateLitigationHold(orgDID: string, body: LitigationHoldRequest) {
  return complyFetch<LitigationMatter>("/comply/litigation/hold", orgDID, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function listEDiscoveryExports(orgDID: string) {
  return complyFetch<{ exports: EDiscoveryExport[] }>("/comply/ediscovery/export", orgDID);
}

export function requestEDiscoveryExport(orgDID: string, body: EDiscoveryExportRequest) {
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

export function auditReportDownloadURL() {
  const q = new URLSearchParams({ format: "pdf" });
  return `${BASE}/comply/audit/report?${q}`;
}
