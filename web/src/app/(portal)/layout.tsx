import { redirect } from "next/navigation";
import { getActiveOrg, getOrgMemberships } from "@/lib/auth/org";
import { PortalShell } from "@/components/portal/PortalShell";

export default async function PortalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const memberships = await getOrgMemberships();
  if (memberships.length === 0) redirect("/login");

  const active = (await getActiveOrg()) ?? memberships[0];

  return (
    <PortalShell memberships={memberships} activeOrgDID={active.orgDID}>
      {children}
    </PortalShell>
  );
}
