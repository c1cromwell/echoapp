import { createClient } from "@/lib/supabase/server";
import { getActiveOrgDIDFromCookie } from "@/lib/auth/active-org";

export type PortalRole = "owner" | "admin" | "compliance_officer" | "auditor" | "viewer";

export interface OrgMembership {
  orgDID: string;
  orgName: string;
  role: PortalRole;
}

/**
 * Resolves the signed-in operator's organization memberships from the portal's
 * Supabase `org_members` table (operator <-> orgDID + portal role). This is the
 * ONLY org/role state the portal owns; the orgDID then scopes every Comply API call.
 *
 * NOTE: roles here are *portal* roles (who can see/configure what in the console).
 * They are distinct from the on-chain EchoOrgRoleCredential (owner/admin/moderator/member),
 * which the Comply backend remains authoritative for.
 */
export async function getOrgMemberships(): Promise<OrgMembership[]> {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) return [];

  const { data, error } = await supabase
    .from("org_members")
    .select("org_did, role, organizations(name)")
    .eq("user_id", user.id)
    .eq("active", true);

  if (error || !data) return [];

  return data.map((row) => ({
    orgDID: row.org_did as string,
    orgName:
      (row as { organizations?: { name?: string } }).organizations?.name ?? row.org_did,
    role: row.role as PortalRole,
  }));
}

/** Returns the active membership (cookie) or the first membership. */
export async function getActiveOrg(orgDID?: string): Promise<OrgMembership | null> {
  const memberships = await getOrgMemberships();
  if (memberships.length === 0) return null;
  const cookieOrg = orgDID ?? (await getActiveOrgDIDFromCookie());
  if (cookieOrg) {
    const hit = memberships.find((m) => m.orgDID === cookieOrg);
    if (hit) return hit;
  }
  return memberships[0];
}
