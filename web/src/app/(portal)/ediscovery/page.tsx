import { getActiveOrg } from "@/lib/auth/org";
import {
  listEDiscoveryExports,
  listLitigationHolds,
  requestEDiscoveryExport,
} from "@/lib/api/comply";
import { ExportRequestForm } from "./ExportRequestForm";

export default async function EDiscoveryPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  const [{ exports }, { matters }] = await Promise.all([
    listEDiscoveryExports(org.orgDID).catch(() => ({ exports: [] })),
    listLitigationHolds(org.orgDID, "all").catch(() => ({ matters: [] })),
  ]);

  async function requestExport(formData: FormData) {
    "use server";
    const active = await getActiveOrg();
    if (!active) return;
    await requestEDiscoveryExport(active.orgDID, {
      matterId: String(formData.get("matter_id") ?? ""),
      custodianSet: String(formData.get("custodian_set") ?? "")
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean),
    });
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold text-white">eDiscovery &amp; matters</h1>
        <p className="mt-1 text-sm text-glacial-muted">
          WO-312 — export jobs, matter registry, and privilege log (conversation-level).
        </p>
      </div>

      <ExportRequestForm action={requestExport} matters={matters} />

      <section>
        <h2 className="text-sm font-semibold text-white">Matters</h2>
        <ul className="mt-3 space-y-2">
          {matters.length === 0 && (
            <li className="text-xs text-glacial-muted">No matters — activate a hold on Retention first.</li>
          )}
          {matters.map((m) => (
            <li
              key={m.matterId}
              className="rounded-xl border border-glacial-border bg-glacial-surface px-4 py-3 text-sm text-white"
            >
              <span className="font-medium">{m.matterId}</span>
              <span className="ml-2 text-xs uppercase text-glacial-muted">{m.status}</span>
              {m.scopeLabel && (
                <p className="mt-1 text-xs text-glacial-muted">{m.scopeLabel}</p>
              )}
            </li>
          ))}
        </ul>
      </section>

      <section>
        <h2 className="text-sm font-semibold text-white">Export jobs</h2>
        <ul className="mt-3 space-y-2">
          {exports.length === 0 && (
            <li className="text-xs text-glacial-muted">No exports yet.</li>
          )}
          {exports.map((e) => (
            <li
              key={e.exportId}
              className="rounded-xl border border-glacial-border bg-glacial-surface px-4 py-3 text-sm"
            >
              <div className="flex justify-between text-white">
                <span>{(e.exportId ?? "unknown").slice(0, 8)}…</span>
                <span className="text-xs uppercase text-glacial-muted">{e.status}</span>
              </div>
              <p className="mt-1 text-xs text-glacial-muted">
                Matter {e.matterId} · {e.messageCount ?? 0} messages
              </p>
              <p className="mt-1 font-mono text-[10px] text-glacial-muted">
                query {e.queryHash?.slice(0, 16) ?? "—"} · checksum{" "}
                {e.dataL1Ref?.slice(0, 16) ?? "—"}
              </p>
            </li>
          ))}
        </ul>
      </section>

      <section className="rounded-2xl border border-glacial-border bg-glacial-surface p-5">
        <h2 className="text-sm font-semibold text-white">Privilege log (FRCP)</h2>
        <p className="mt-2 text-xs text-glacial-muted">
          Conversation-level privilege designations and ethical-wall overrides land here (WO-262–264).
          Privileged threads are excluded from exports by default; overrides are audit-logged with
          Data L1 refs.
        </p>
        <p className="mt-3 text-xs text-glacial-muted">No privileged conversations designated yet.</p>
      </section>
    </div>
  );
}
