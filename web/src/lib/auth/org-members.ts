import "server-only";

import { createClient } from "@/lib/supabase/server";
import { scimAdminClient } from "@/lib/scim/auth";
import type { PortalRole } from "@/lib/auth/org";

export interface PortalMemberRow {
  userId: string;
  email: string;
  role: PortalRole;
  active: boolean;
  scimExternalId?: string;
}

/** Lists portal operator seats for an org (WO-310). Requires owner/admin portal role. */
export async function listPortalMembers(orgDID: string): Promise<PortalMemberRow[]> {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();
  if (!user) return [];

  const { data: self } = await supabase
    .from("org_members")
    .select("role")
    .eq("user_id", user.id)
    .eq("org_did", orgDID)
    .eq("active", true)
    .maybeSingle();

  if (!self || !["owner", "admin"].includes(self.role as string)) {
    return [];
  }

  const admin = scimAdminClient();
  const { data: rows } = await admin
    .from("org_members")
    .select("user_id, role, active, scim_external_id")
    .eq("org_did", orgDID)
    .order("created_at", { ascending: true });

  if (!rows) return [];

  const members: PortalMemberRow[] = [];
  for (const row of rows) {
    const { data: authUser } = await admin.auth.admin.getUserById(row.user_id as string);
    members.push({
      userId: row.user_id as string,
      email: authUser.user?.email ?? "(unknown)",
      role: row.role as PortalRole,
      active: row.active as boolean,
      scimExternalId: (row.scim_external_id as string) || undefined,
    });
  }
  return members;
}

export async function countActivePortalSeats(orgDID: string): Promise<number> {
  const admin = scimAdminClient();
  const { count } = await admin
    .from("org_members")
    .select("*", { count: "exact", head: true })
    .eq("org_did", orgDID)
    .eq("active", true);
  return count ?? 0;
}
