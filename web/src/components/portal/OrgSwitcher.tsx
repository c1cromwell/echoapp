"use client";

import { useTransition } from "react";
import { useRouter } from "next/navigation";
import type { OrgMembership } from "@/lib/auth/org";
import { setActiveOrgAction, signOutAction } from "@/lib/auth/session";

export function OrgSwitcher({
  memberships,
  activeOrgDID,
}: {
  memberships: OrgMembership[];
  activeOrgDID: string;
}) {
  const router = useRouter();
  const [pending, startTransition] = useTransition();

  function onOrgChange(orgDID: string) {
    startTransition(async () => {
      await setActiveOrgAction(orgDID);
      router.refresh();
    });
  }

  const active = memberships.find((m) => m.orgDID === activeOrgDID) ?? memberships[0];

  return (
    <div className="mb-6">
      <p className="text-sm font-semibold text-white">Echo Comply</p>
      <select
        value={activeOrgDID}
        disabled={pending}
        onChange={(e) => onOrgChange(e.target.value)}
        className="mt-2 w-full rounded-lg border border-glacial-border bg-glacial-bg px-2 py-1 text-xs text-glacial-muted"
      >
        {memberships.map((m) => (
          <option key={m.orgDID} value={m.orgDID}>
            {m.orgName}
          </option>
        ))}
      </select>
      <p className="mt-1 truncate text-[10px] text-glacial-muted" title={active?.orgDID}>
        {active?.orgDID}
      </p>
      <p className="mt-2 text-[10px] text-glacial-muted">Role: {active?.role}</p>
      <form action={signOutAction} className="mt-4">
        <button
          type="submit"
          className="w-full rounded-lg border border-glacial-border px-2 py-1 text-xs text-glacial-muted hover:text-white"
        >
          Sign out
        </button>
      </form>
    </div>
  );
}
