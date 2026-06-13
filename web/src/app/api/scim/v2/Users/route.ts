import { NextResponse } from "next/server";
import { isAuthorizedScim } from "@/lib/scim/auth";

/**
 * SCIM 2.0 Users endpoint (RFC 7644) — enterprise IdP provisioning of portal seats.
 *
 * WO-309 ships the authenticated contract surface (auth gate + SCIM-shaped responses).
 * The full provisioning logic (upsert into Supabase `org_members`, map IdP groups -> portal
 * roles, soft-deactivate on DELETE) lands with WO-310. Mapping IdP group -> orgDID is a
 * per-tenant config.
 */

const SCIM_CONTENT_TYPE = "application/scim+json";

function unauthorized() {
  return NextResponse.json(
    { schemas: ["urn:ietf:params:scim:api:messages:2.0:Error"], status: "401", detail: "Unauthorized" },
    { status: 401, headers: { "Content-Type": SCIM_CONTENT_TYPE } },
  );
}

export async function GET(request: Request) {
  if (!isAuthorizedScim(request)) return unauthorized();
  // ListResponse shell — WO-310 backs this with real `org_members` queries + filtering.
  return NextResponse.json(
    {
      schemas: ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
      totalResults: 0,
      startIndex: 1,
      itemsPerPage: 0,
      Resources: [],
    },
    { headers: { "Content-Type": SCIM_CONTENT_TYPE } },
  );
}

export async function POST(request: Request) {
  if (!isAuthorizedScim(request)) return unauthorized();
  const body = await request.json().catch(() => ({}));
  const userName: string = body.userName ?? body.emails?.[0]?.value ?? "";
  if (!userName) {
    return NextResponse.json(
      { schemas: ["urn:ietf:params:scim:api:messages:2.0:Error"], status: "400", detail: "userName required" },
      { status: 400, headers: { "Content-Type": SCIM_CONTENT_TYPE } },
    );
  }

  // TODO(WO-310): upsert operator + org_members row via scimAdminClient(); map groups -> role.
  return NextResponse.json(
    {
      schemas: ["urn:ietf:params:scim:schemas:core:2.0:User"],
      id: crypto.randomUUID(),
      userName,
      active: true,
      meta: { resourceType: "User" },
    },
    { status: 201, headers: { "Content-Type": SCIM_CONTENT_TYPE } },
  );
}
