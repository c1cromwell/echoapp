import { getActiveOrg } from "@/lib/auth/org";
import {
  activateLitigationHold,
  createRetentionPolicy,
  listLitigationHolds,
  listRetentionPolicies,
  releaseLitigationHold,
  type RetentionPolicyCreate,
} from "@/lib/api/comply";
import { RetentionPolicyForm } from "./RetentionPolicyForm";
import { ReleaseHoldButton } from "./ReleaseHoldButton";

function anchorBadge(ref?: string) {
  if (!ref) return "pending";
  return ref.length > 20 ? `${ref.slice(0, 12)}…` : ref;
}

export default async function RetentionPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  const [{ policies }, { matters }] = await Promise.all([
    listRetentionPolicies(org.orgDID).catch(() => ({ policies: [] })),
    listLitigationHolds(org.orgDID).catch(() => ({ matters: [] })),
  ]);

  async function createPolicy(formData: FormData) {
    "use server";
    const active = await getActiveOrg();
    if (!active) return;
    await createRetentionPolicy(active.orgDID, {
      policy_type: String(formData.get("policy_type") ?? "permanent") as RetentionPolicyCreate["policy_type"],
      conversation_id: String(formData.get("conversation_id") ?? "") || undefined,
      scope_label: String(formData.get("scope_label") ?? "") || undefined,
    });
  }

  async function createHold(formData: FormData) {
    "use server";
    const active = await getActiveOrg();
    if (!active) return;
    const custodians = String(formData.get("custodian_dids") ?? "")
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    await activateLitigationHold(active.orgDID, {
      matterId: String(formData.get("matter_id") ?? ""),
      custodianDids: custodians,
      scope: String(formData.get("scope") ?? "") || undefined,
    });
  }

  async function releaseHold(formData: FormData) {
    "use server";
    const active = await getActiveOrg();
    if (!active) return;
    const matterId = String(formData.get("matter_id") ?? "");
    if (!matterId) return;
    await releaseLitigationHold(active.orgDID, matterId);
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold text-white">Retention &amp; litigation holds</h1>
        <p className="mt-1 text-sm text-glacial-muted">
          WO-311 — policies and holds with Data L1 anchor confirmations (hash refs only).
        </p>
      </div>

      <RetentionPolicyForm action={createPolicy} holdAction={createHold} />

      <section>
        <h2 className="text-sm font-semibold text-white">Active policies</h2>
        <ul className="mt-3 space-y-2">
          {policies.length === 0 && (
            <li className="text-xs text-glacial-muted">No policies yet.</li>
          )}
          {policies.map((p) => (
            <li
              key={p.id}
              className="rounded-xl border border-glacial-border bg-glacial-surface px-4 py-3 text-sm text-white"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="font-medium">{p.policyType}</span>
                <span className="font-mono text-[10px] text-emerald-400" title={p.dataL1Ref}>
                  anchor {anchorBadge(p.dataL1Ref)}
                </span>
              </div>
              {p.conversationId && (
                <p className="mt-1 text-xs text-glacial-muted">conv {p.conversationId}</p>
              )}
              {p.scopeLabel && (
                <p className="text-xs text-glacial-muted">{p.scopeLabel}</p>
              )}
            </li>
          ))}
        </ul>
      </section>

      <section>
        <h2 className="text-sm font-semibold text-white">Litigation holds</h2>
        <ul className="mt-3 space-y-2">
          {matters.length === 0 && (
            <li className="text-xs text-glacial-muted">No active holds.</li>
          )}
          {matters.map((m) => (
            <li
              key={m.matterId}
              className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-glacial-border bg-glacial-surface px-4 py-3 text-sm"
            >
              <div>
                <p className="font-medium text-white">{m.matterId}</p>
                <p className="text-xs text-glacial-muted">
                  {m.custodianCount ?? 0} custodians · {m.status} · anchor{" "}
                  <span className="font-mono text-emerald-400">{anchorBadge(m.dataL1Ref)}</span>
                </p>
              </div>
              {m.status === "active" && m.matterId && (
                <ReleaseHoldButton matterId={m.matterId} action={releaseHold} />
              )}
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
