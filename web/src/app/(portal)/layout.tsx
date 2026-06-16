import Link from "next/link";
import { redirect } from "next/navigation";
import { getActiveOrg, getOrgMemberships } from "@/lib/auth/org";
import { OrgSwitcher } from "@/components/portal/OrgSwitcher";

const NAV = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/organization", label: "Organization" },
  { href: "/retention", label: "Retention & Holds" },
  { href: "/ediscovery", label: "eDiscovery & Matters" },
];

export default async function PortalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const memberships = await getOrgMemberships();
  if (memberships.length === 0) redirect("/login");

  const active = (await getActiveOrg()) ?? memberships[0];

  return (
    <div className="flex min-h-screen">
      <aside className="w-64 shrink-0 border-r border-glacial-border bg-glacial-surface p-4">
        <OrgSwitcher memberships={memberships} activeOrgDID={active.orgDID} />

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
      </aside>

      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}
