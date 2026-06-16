import { cookies } from "next/headers";

export const ACTIVE_ORG_COOKIE = "echo_active_org";

export async function getActiveOrgDIDFromCookie(): Promise<string | undefined> {
  const jar = await cookies();
  return jar.get(ACTIVE_ORG_COOKIE)?.value;
}
