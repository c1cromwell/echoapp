import { createClient } from "@supabase/supabase-js";

/** Bearer-token check for SCIM endpoints (machine-to-machine; the IdP presents this token). */
export function isAuthorizedScim(request: Request): boolean {
  const auth = request.headers.get("authorization") ?? "";
  const token = auth.replace(/^Bearer\s+/i, "");
  const expected = process.env.SCIM_BEARER_TOKEN;
  return Boolean(expected) && token === expected;
}

/**
 * Service-role Supabase client for provisioning (server-only). Used to upsert/deactivate
 * operator seats + `org_members` rows from the customer's IdP. Never imported by client code.
 */
export function scimAdminClient() {
  return createClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.SUPABASE_SERVICE_ROLE_KEY!,
    { auth: { persistSession: false } },
  );
}
