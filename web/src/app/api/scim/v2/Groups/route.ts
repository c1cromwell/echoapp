import { NextResponse } from "next/server";
import { isAuthorizedScim } from "@/lib/scim/auth";
import { SCIM_CONTENT_TYPE, scimError } from "@/lib/scim/provision";

const PORTAL_GROUPS = [
  { id: "echo-comply-owner", displayName: "Echo-Comply-Owner" },
  { id: "echo-comply-admin", displayName: "Echo-Comply-Admin" },
  { id: "echo-comply-compliance", displayName: "Echo-Comply-Compliance" },
  { id: "echo-comply-auditor", displayName: "Echo-Comply-Auditor" },
  { id: "echo-comply-viewer", displayName: "Echo-Comply-Viewer" },
];

export async function GET(request: Request) {
  if (!isAuthorizedScim(request)) return scimError(401, "Unauthorized");

  const resources = PORTAL_GROUPS.map((g) => ({
    schemas: ["urn:ietf:params:scim:schemas:core:2.0:Group"],
    id: g.id,
    displayName: g.displayName,
    meta: { resourceType: "Group" },
  }));

  return NextResponse.json(
    {
      schemas: ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
      totalResults: resources.length,
      startIndex: 1,
      itemsPerPage: resources.length,
      Resources: resources,
    },
    { headers: { "Content-Type": SCIM_CONTENT_TYPE } },
  );
}
