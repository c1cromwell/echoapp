import { getActiveOrg } from "@/lib/auth/org";
import { listEDiscoveryExports, requestEDiscoveryExport } from "@/lib/api/comply";
import { ExportRequestForm } from "./ExportRequestForm";

export default async function EDiscoveryPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  const { exports } = await listEDiscoveryExports(org.orgDID).catch(() => ({ exports: [] }));

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
          WO-312 — export requests with Merkle refs and Data L1 checksum anchors (metadata only).
        </p>
      </div>

      <ExportRequestForm action={requestExport} />

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
                Matter {e.matterId} · {e.messageCount} messages · ref {e.dataL1Ref?.slice(0, 12) ?? "—"}
              </p>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
