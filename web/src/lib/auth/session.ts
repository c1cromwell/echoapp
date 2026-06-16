"use server";

import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { ACTIVE_ORG_COOKIE } from "@/lib/auth/active-org";
import { cookies } from "next/headers";

export async function signOutAction() {
  const supabase = await createClient();
  await supabase.auth.signOut();
  const jar = await cookies();
  jar.delete(ACTIVE_ORG_COOKIE);
  redirect("/login");
}

export async function setActiveOrgAction(orgDID: string) {
  const jar = await cookies();
  jar.set(ACTIVE_ORG_COOKIE, orgDID, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 60 * 24 * 365,
  });
}
