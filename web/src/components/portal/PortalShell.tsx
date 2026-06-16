"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { OrgMembership } from "@/lib/auth/org";
import { OrgSwitcher } from "@/components/portal/OrgSwitcher";

const NAV = [
  { href: "/dashboard", label: "Dashboard", short: "Home" },
  { href: "/organization", label: "Organization", short: "Org" },
  { href: "/retention", label: "Retention & Holds", short: "Holds" },
  { href: "/ediscovery", label: "eDiscovery", short: "Export" },
];

export function PortalShell({
  memberships,
  activeOrgDID,
  children,
}: {
  memberships: OrgMembership[];
  activeOrgDID: string;
  children: React.ReactNode;
}) {
  const pathname = usePathname();

  return (
    <div className="flex min-h-screen flex-col md:flex-row">
      <aside className="hidden w-64 shrink-0 border-r border-glacial-border bg-glacial-surface p-4 md:block">
        <OrgSwitcher memberships={memberships} activeOrgDID={activeOrgDID} />
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

      <div className="border-b border-glacial-border bg-glacial-surface p-4 md:hidden">
        <OrgSwitcher memberships={memberships} activeOrgDID={activeOrgDID} />
      </div>

      <main className="flex-1 p-4 pb-24 md:p-8 md:pb-8">{children}</main>

      <nav
        className="fixed inset-x-0 bottom-0 z-10 flex border-t border-glacial-border bg-glacial-surface md:hidden"
        aria-label="Mobile navigation"
      >
        {NAV.map((item) => {
          const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex flex-1 flex-col items-center py-2 text-[10px] ${
                active ? "text-white" : "text-glacial-muted"
              }`}
            >
              <span className="font-medium">{item.short}</span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
