import { getActiveOrg } from "@/lib/auth/org";
import { countActivePortalSeats, listPortalMembers } from "@/lib/auth/org-members";
import { getOrgProfile } from "@/lib/api/comply";

export default async function OrganizationPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  const [profile, members, seatCount] = await Promise.all([
    getOrgProfile(org.orgDID).catch(() => null),
    listPortalMembers(org.orgDID),
    countActivePortalSeats(org.orgDID),
  ]);

  const seats = profile?.seats ?? 10;
  const atLimit = seatCount >= seats;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-semibold text-white">Organization</h1>
        <p className="mt-1 text-sm text-glacial-muted">
          Portal seats, tier limits, and operator roles (WO-310). On-chain{" "}
          <code className="text-xs">EchoOrgRoleCredential</code> sync is backend-owned.
        </p>
      </div>

      <dl className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[
          { label: "Org name", value: org.orgName },
          { label: "Tier", value: profile?.tier ?? "starter" },
          { label: "Seats used", value: `${seatCount} / ${seats}` },
          { label: "Your role", value: org.role },
        ].map((item) => (
          <div
            key={item.label}
            className="rounded-2xl border border-glacial-border bg-glacial-surface p-4"
          >
            <dt className="text-xs text-glacial-muted">{item.label}</dt>
            <dd className="mt-1 text-sm font-medium text-white">{item.value}</dd>
          </div>
        ))}
      </dl>

      {atLimit && (
        <p className="rounded-xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-xs text-amber-200">
          Seat limit reached — provision additional operators via SCIM or upgrade tier.
        </p>
      )}

      <section>
        <h2 className="text-sm font-semibold text-white">Portal operators</h2>
        <p className="mt-1 text-xs text-glacial-muted">
          Operator emails live in Supabase auth only; Comply dashboards never receive them.
        </p>
        <div className="mt-3 overflow-x-auto rounded-2xl border border-glacial-border">
          <table className="min-w-full text-left text-xs">
            <thead className="bg-glacial-surface text-glacial-muted">
              <tr>
                <th className="px-4 py-2 font-medium">Email</th>
                <th className="px-4 py-2 font-medium">Portal role</th>
                <th className="px-4 py-2 font-medium">Status</th>
                <th className="px-4 py-2 font-medium">SCIM ID</th>
              </tr>
            </thead>
            <tbody>
              {members.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-4 text-glacial-muted">
                    {["owner", "admin"].includes(org.role)
                      ? "No members yet — use SCIM or invite flow."
                      : "Admin view only (owner/admin role required)."}
                  </td>
                </tr>
              )}
              {members.map((m) => (
                <tr key={m.userId} className="border-t border-glacial-border text-white">
                  <td className="px-4 py-2">{m.email}</td>
                  <td className="px-4 py-2">{m.role}</td>
                  <td className="px-4 py-2">{m.active ? "active" : "inactive"}</td>
                  <td className="px-4 py-2 font-mono text-[10px] text-glacial-muted">
                    {m.scimExternalId?.slice(0, 12) ?? "—"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="rounded-2xl border border-glacial-border bg-glacial-surface p-5">
        <h2 className="text-sm font-semibold text-white">Org DID</h2>
        <p className="mt-2 font-mono text-xs text-glacial-muted">{org.orgDID}</p>
      </section>
    </div>
  );
}
