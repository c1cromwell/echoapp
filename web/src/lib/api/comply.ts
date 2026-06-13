import "server-only";

/**
 * Server-only client for the Go Comply backend (system of record).
 *
 * The portal never talks to the Comply API from the browser and never caches its
 * payloads in Supabase. Every call is scoped by `orgDID` via the `X-Org-DID` header
 * and authenticated with a service token. Once the WO-250+ endpoints + schemas land,
 * run `npm run generate-api` and replace the hand-typed shapes below with the generated
 * types from `./schema.d.ts`.
 */

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
    // Compliance data is per-request and must never be statically cached.
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Comply API ${path} -> ${res.status}`);
  }
  return (await res.json()) as T;
}

/** Shape mirrors WO-252 `GET /comply/dashboard`. Zero-PII: counts/rates/health only. */
export interface ComplyDashboardSummary {
  deCoverageRate: string;
  activeRetentionPolicies: number;
  litigationHolds: number;
  pendingExports: number;
  anchorHealth: "healthy" | "degraded" | "down";
}

export function getComplyDashboard(orgDID: string) {
  return complyFetch<ComplyDashboardSummary>("/comply/dashboard", orgDID);
}
