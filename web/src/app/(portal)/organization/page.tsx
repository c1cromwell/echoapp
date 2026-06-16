import { getActiveOrg } from "@/lib/auth/org";

export default async function OrganizationPage() {
  const org = await getActiveOrg();
  if (!org) return null;

  return (
    <div>
      <h1 className="text-2xl font-semibold text-white">Organization</h1>
      <p className="mt-1 text-sm text-glacial-muted">
        WO-310 — member invites, seat limits, and on-chain role credentials land here.
      </p>
      <dl className="mt-6 space-y-3 rounded-2xl border border-glacial-border bg-glacial-surface p-5 text-sm">
        <div>
          <dt className="text-glacial-muted">Name</dt>
          <dd className="text-white">{org.orgName}</dd>
        </div>
        <div>
          <dt className="text-glacial-muted">Org DID</dt>
          <dd className="font-mono text-xs text-white">{org.orgDID}</dd>
        </div>
        <div>
          <dt className="text-glacial-muted">Your portal role</dt>
          <dd className="text-white">{org.role}</dd>
        </div>
      </dl>
    </div>
  );
}
