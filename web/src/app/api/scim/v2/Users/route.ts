import { NextResponse } from "next/server";
import { isAuthorizedScim } from "@/lib/scim/auth";
import {
  SCIM_CONTENT_TYPE,
  listScimUsers,
  provisionScimUser,
  scimError,
} from "@/lib/scim/provision";

function unauthorized() {
  return scimError(401, "Unauthorized");
}

export async function GET(request: Request) {
  if (!isAuthorizedScim(request)) return unauthorized();
  const orgDID = process.env.SCIM_DEFAULT_ORG_DID?.trim();
  if (!orgDID) return scimError(503, "SCIM_DEFAULT_ORG_DID not configured");

  const users = await listScimUsers(orgDID);
  return NextResponse.json(
    {
      schemas: ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
      totalResults: users.length,
      startIndex: 1,
      itemsPerPage: users.length,
      Resources: users,
    },
    { headers: { "Content-Type": SCIM_CONTENT_TYPE } },
  );
}

export async function POST(request: Request) {
  if (!isAuthorizedScim(request)) return unauthorized();
  const body = await request.json().catch(() => ({}));
  try {
    const user = await provisionScimUser(body as Record<string, unknown>);
    return NextResponse.json(user, {
      status: 201,
      headers: { "Content-Type": SCIM_CONTENT_TYPE },
    });
  } catch (err) {
    const msg = err instanceof Error ? err.message : "provision failed";
    return scimError(400, msg);
  }
}
