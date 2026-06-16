import { NextResponse } from "next/server";
import { isAuthorizedScim } from "@/lib/scim/auth";
import {
  SCIM_CONTENT_TYPE,
  deactivateScimUser,
  provisionScimUser,
  scimError,
  toScimUser,
} from "@/lib/scim/provision";
import { scimAdminClient } from "@/lib/scim/auth";

type RouteContext = { params: Promise<{ id: string }> };

export async function GET(_request: Request, context: RouteContext) {
  if (!isAuthorizedScim(_request)) return scimError(401, "Unauthorized");
  const { id } = await context.params;
  const admin = scimAdminClient();
  const { data } = await admin
    .from("org_members")
    .select("user_id, active, role")
    .eq("scim_external_id", id)
    .maybeSingle();
  if (!data) return scimError(404, "User not found");
  const { data: authUser } = await admin.auth.admin.getUserById(data.user_id as string);
  const email = authUser.user?.email ?? id;
  return NextResponse.json(
    toScimUser(id, email, data.active as boolean, [{ display: data.role as string }]),
    { headers: { "Content-Type": SCIM_CONTENT_TYPE } },
  );
}

export async function PATCH(request: Request, context: RouteContext) {
  if (!isAuthorizedScim(request)) return scimError(401, "Unauthorized");
  const { id } = await context.params;
  const body = await request.json().catch(() => ({}));
  if (body.active === false) {
    await deactivateScimUser(id);
    return NextResponse.json(
      toScimUser(id, id, false),
      { headers: { "Content-Type": SCIM_CONTENT_TYPE } },
    );
  }
  try {
    const user = await provisionScimUser({ ...body, id });
    return NextResponse.json(user, { headers: { "Content-Type": SCIM_CONTENT_TYPE } });
  } catch (err) {
    const msg = err instanceof Error ? err.message : "update failed";
    return scimError(400, msg);
  }
}

export async function DELETE(request: Request, context: RouteContext) {
  if (!isAuthorizedScim(request)) return scimError(401, "Unauthorized");
  const { id } = await context.params;
  await deactivateScimUser(id);
  return new NextResponse(null, { status: 204 });
}
