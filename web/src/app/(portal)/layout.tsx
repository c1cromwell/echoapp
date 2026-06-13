import Link from "next/link";
import { redirect } from "next/navigation";
import { getOrgMemberships } from "@/lib/auth/org";

const NAV = [
  { href: "/dashboard", label: "Dashboard", wo: "WO-313" },
  { href: "/organization", label: "Organization", wo: "WO-310" },
  { href: "/retention", label: "Retention & Holds", wo: "WO-311" },
  { href: "/ediscovery", label: "eDiscovery & Matters", wo: "WO-312" },
];

/**
 * Authenticated portal shell. Every page below is org-scoped: if the operator has
 * no org membership we send them back to login (no orphan sessions).
 */
export default async function PortalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const memberships = await getOrgMemberships();
  if (memberships.length === 0) redirect("/login");

  const active = memberships[0];

  return (
    <div className="flex min-h-screen">
      <aside className="w-64 shrink-0 border-r border-glacial-border bg-glacial-surface p-4">
        <div className="mb-6">
          <p className="text-sm font-semibold text-white">Echo Comply</p>
          {/* Org switcher (WO-310 expands this into a real selector) */}
          <select
            defaultValue={active.orgDID}
            className="mt-2 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-1 text-xs text-glacial-muted"
          >
            {memberships.map((m) => (
              <option key={m.orgDID} value={m.orgDID}>
                {m.orgName}
              </option>
            ))}
          </select>
          <p className="mt-1 truncate text-[10px] text-glacial-muted" title={active.orgDID}>
            {active.orgDID}
          </p>
        </div>

        <nav className="space-y-1">
          {NAV.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="block rounded-lg px-3 py-2 text-sm text-glacial-muted hover:bg-glacial-bg hover:text-white"
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <p className="mt-8 text-[10px] text-glacial-muted">
          Role: {active.role}
        </p>
      </aside>

      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}
