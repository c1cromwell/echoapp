import { NextResponse } from "next/server";
import { isAuthorizedScim } from "@/lib/scim/auth";

/**
 * SCIM 2.0 Groups endpoint (RFC 7644). Groups map to portal roles / orgDID scope.
 * WO-309 ships the auth-gated contract surface; membership sync lands with WO-310.
 */

const SCIM_CONTENT_TYPE = "application/scim+json";

export async function GET(request: Request) {
  if (!isAuthorizedScim(request)) {
    return NextResponse.json(
      { schemas: ["urn:ietf:params:scim:api:messages:2.0:Error"], status: "401", detail: "Unauthorized" },
      { status: 401, headers: { "Content-Type": SCIM_CONTENT_TYPE } },
    );
  }
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
