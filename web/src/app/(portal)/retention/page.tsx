import { getActiveOrg } from "@/lib/auth/org";
import {
  activateLitigationHold,
  createRetentionPolicy,
  listRetentionPolicies,
  type RetentionPolicyCreate,
} from "@/lib/api/comply";
import { RetentionPolicyForm } from "./RetentionPolicyForm";

export default async function RetentionPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  const { policies } = await listRetentionPolicies(org.orgDID).catch(() => ({ policies: [] }));

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

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold text-white">Retention &amp; litigation holds</h1>
        <p className="mt-1 text-sm text-glacial-muted">
          WO-311 — configure org policies and activate holds (server-side enforcement only).
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
              <span className="font-medium">{p.policyType}</span>
              {p.conversationId && (
                <span className="ml-2 text-xs text-glacial-muted">conv {p.conversationId}</span>
              )}
              {p.scopeLabel && (
                <span className="ml-2 text-xs text-glacial-muted">{p.scopeLabel}</span>
              )}
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
