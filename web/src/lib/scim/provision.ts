import type { PortalRole } from "@/lib/auth/org";
import { scimAdminClient } from "@/lib/scim/auth";

const SCIM_CONTENT_TYPE = "application/scim+json";

export { SCIM_CONTENT_TYPE };

export interface ScimGroupRef {
  value?: string;
  display?: string;
  "$ref"?: string;
}

export interface ScimUserResource {
  schemas: string[];
  id: string;
  userName: string;
  active: boolean;
  emails?: { value: string; primary?: boolean }[];
  groups?: ScimGroupRef[];
  meta: { resourceType: string; location?: string };
}

const DEFAULT_GROUP_MAP: Record<string, PortalRole> = {
  "Echo-Comply-Owner": "owner",
  "Echo-Comply-Admin": "admin",
  "Echo-Comply-Compliance": "compliance_officer",
  "Echo-Comply-Auditor": "auditor",
  "Echo-Comply-Viewer": "viewer",
};

function groupRoleMap(): Record<string, PortalRole> {
  const raw = process.env.SCIM_GROUP_ROLE_MAP;
  if (!raw) return DEFAULT_GROUP_MAP;
  try {
    return { ...DEFAULT_GROUP_MAP, ...JSON.parse(raw) };
  } catch {
    return DEFAULT_GROUP_MAP;
  }
}

export function resolveScimOrgDID(body: Record<string, unknown>): string {
  const fromEnv = process.env.SCIM_DEFAULT_ORG_DID?.trim();
  const ext = body["urn:echo:params:scim:schemas:extension:org:2.0:User"] as
    | { orgDid?: string }
    | undefined;
  return (ext?.orgDid ?? fromEnv ?? "").trim();
}

export function mapScimGroupsToRole(groups?: ScimGroupRef[]): PortalRole {
  const map = groupRoleMap();
  for (const g of groups ?? []) {
    const key = g.display ?? g.value ?? "";
    if (key && map[key]) return map[key];
  }
  return "viewer";
}

export async function ensureOrganization(orgDID: string, name?: string) {
  const admin = scimAdminClient();
  const orgName = name ?? process.env.SCIM_DEFAULT_ORG_NAME ?? orgDID;
  await admin.from("organizations").upsert({ org_did: orgDID, name: orgName });
}

async function findUserIdByEmail(email: string): Promise<string | null> {
  const admin = scimAdminClient();
  const { data, error } = await admin.auth.admin.listUsers({ page: 1, perPage: 200 });
  if (error || !data?.users) return null;
  const hit = data.users.find((u) => u.email?.toLowerCase() === email.toLowerCase());
  return hit?.id ?? null;
}

export async function provisionScimUser(body: Record<string, unknown>): Promise<ScimUserResource> {
  const userName = String(body.userName ?? (body.emails as { value: string }[])?.[0]?.value ?? "");
  if (!userName.includes("@")) {
    throw new Error("userName must be an email");
  }
  const orgDID = resolveScimOrgDID(body);
  if (!orgDID) {
    throw new Error("SCIM_DEFAULT_ORG_DID or org extension required");
  }

  const scimId = String(body.id ?? crypto.randomUUID());
  const active = body.active !== false;
  const role = mapScimGroupsToRole(body.groups as ScimGroupRef[] | undefined);

  await ensureOrganization(orgDID);

  const admin = scimAdminClient();
  let userId = await findUserIdByEmail(userName);
  if (!userId) {
    const { data, error } = await admin.auth.admin.createUser({
      email: userName,
      email_confirm: true,
    });
    if (error || !data.user) throw new Error(error?.message ?? "create user failed");
    userId = data.user.id;
  }

  await admin.from("org_members").upsert(
    {
      user_id: userId,
      org_did: orgDID,
      role,
      scim_external_id: scimId,
      active,
    },
    { onConflict: "user_id,org_did" },
  );

  return toScimUser(scimId, userName, active, body.groups as ScimGroupRef[] | undefined);
}

export async function listScimUsers(orgDID: string): Promise<ScimUserResource[]> {
  const admin = scimAdminClient();
  const { data, error } = await admin
    .from("org_members")
    .select("scim_external_id, role, active, user_id")
    .eq("org_did", orgDID)
    .eq("active", true);
  if (error || !data) return [];

  const users: ScimUserResource[] = [];
  for (const row of data) {
    const { data: authUser } = await admin.auth.admin.getUserById(row.user_id as string);
    const email = authUser.user?.email ?? row.user_id;
    users.push(
      toScimUser(
        (row.scim_external_id as string) ?? row.user_id,
        email,
        row.active as boolean,
        [{ display: row.role as string }],
      ),
    );
  }
  return users;
}

export async function deactivateScimUser(scimId: string): Promise<void> {
  const admin = scimAdminClient();
  const { data } = await admin
    .from("org_members")
    .select("user_id, org_did")
    .eq("scim_external_id", scimId)
    .maybeSingle();
  if (!data) return;
  await admin
    .from("org_members")
    .update({ active: false })
    .eq("scim_external_id", scimId);
}

export function toScimUser(
  id: string,
  userName: string,
  active: boolean,
  groups?: ScimGroupRef[],
): ScimUserResource {
  return {
    schemas: ["urn:ietf:params:scim:schemas:core:2.0:User"],
    id,
    userName,
    active,
    emails: [{ value: userName, primary: true }],
    groups,
    meta: { resourceType: "User" },
  };
}

export function scimError(status: number, detail: string) {
  return Response.json(
    {
      schemas: ["urn:ietf:params:scim:api:messages:2.0:Error"],
      status: String(status),
      detail,
    },
    { status, headers: { "Content-Type": SCIM_CONTENT_TYPE } },
  );
}
